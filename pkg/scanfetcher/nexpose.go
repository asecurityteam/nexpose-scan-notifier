package scanfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/container"
	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
)

const (
	activeQueryParam = "active" // Bool indicating if active scans should be fetched.
	pageQueryParam   = "page"   // The index of the page (zero-based) to retrieve.
	sizeQueryParam   = "size"   // The number of records per page to retrieve.
	sortQueryParam   = "sort"
	sortQueryValue   = "endTime,DESC" // Return scans in descending order start with most recently completed.

	finishedScanStatus = "finished" // Status for scans which have completed successfully.
)

type page struct {
	Number         int `json:"number"`
	Size           int `json:"size"`
	TotalPages     int `json:"totalPages"`
	TotalResources int `json:"totalResources"`
}

type resource struct {
	ScanID    int    `json:"id"`
	SiteID    int    `json:"siteId"`
	ScanType  string `json:"scanType"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	ScanName  string `json:"scanName"`
	Status    string `json:"status"`
}

type nexposeScanResponse struct {
	Page      page       `json:"page"`
	Resources []resource `json:"resources"`
}

// NexposeClient implements the interfaces to fetch scans from Nexpose
type NexposeClient struct {
	Client        *http.Client
	Endpoint      *url.URL
	PageSize      int
	ScanBlocklist *container.StringContainer
}

// FetchScans fetches Nexpose scans, filters out running scans, and returns all completed scans
// after the provided timestamp.
func (n *NexposeClient) FetchScans(ctx context.Context, ts time.Time) ([]domain.CompletedScan, error) {
	var completedScans []domain.CompletedScan

	scanResp, err := n.makePagedNexposeScanRequest(0)
	if err != nil {
		return nil, err
	}

	pages := scanResp.Page.TotalPages
	for _, resource := range scanResp.Resources {
		completedScan, err := n.scanResourceToCompletedScan(resource, ts)
		switch err.(type) {
		case nil:
			completedScans = append(completedScans, completedScan)
		case scanNotFinishedError:
			// skip scans without a status of "finished"
		case scanNameInBlocklistError:
			//skip scans included by name in the blocklist
		case outOfRangeError:
			// since scans are returned in descending order by scan time, return
			// the list of completed scans after finding the first scan outside
			// the valid time range
			return completedScans, nil
		default:
			return nil, err
		}
	}

	for curPage := 1; curPage < pages; curPage = curPage + 1 {
		scanResp, err := n.makePagedNexposeScanRequest(curPage)
		if err != nil {
			return nil, err
		}

		for _, resource := range scanResp.Resources {
			completedScan, err := n.scanResourceToCompletedScan(resource, ts)
			switch err.(type) {
			case nil:
				completedScans = append(completedScans, completedScan)
			case scanNotFinishedError:
				// skip scans without a status of "finished"
			case scanNameInBlocklistError:
				//skip scans included by name in the blocklist
			case outOfRangeError:
				// since scans are returned in descending order by scan time, return
				// the list of completed scans after finding the first scan outside
				// the valid time range
				return completedScans, nil
			default:
				return nil, err
			}
		}
	}

	return completedScans, nil
}

func (n *NexposeClient) makePagedNexposeScanRequest(page int) (nexposeScanResponse, error) {
	u, _ := url.Parse(n.Endpoint.String())
	u.Path = path.Join(u.Path, "api", "3", "scans")

	q := u.Query()
	q.Set(activeQueryParam, "false")
	q.Set(pageQueryParam, strconv.Itoa(page))
	q.Set(sizeQueryParam, strconv.Itoa(n.PageSize))
	q.Set(sortQueryParam, sortQueryValue)
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	req.Header.Set("Content-Type", "application/json")
	res, err := n.Client.Do(req)
	if err != nil {
		return nexposeScanResponse{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nexposeScanResponse{}, err
	}

	if res.StatusCode != http.StatusOK {
		return nexposeScanResponse{}, fmt.Errorf("unexpected response from nexpose scans api: %d %s",
			res.StatusCode, string(body))
	}

	var scanResp nexposeScanResponse
	err = json.Unmarshal(body, &scanResp)
	if err != nil {
		return nexposeScanResponse{}, err
	}
	return scanResp, nil
}

func (n *NexposeClient) scanResourceToCompletedScan(resource resource, start time.Time) (domain.CompletedScan, error) {
	// skip scans that have not finished
	if !strings.EqualFold(resource.Status, finishedScanStatus) {
		return domain.CompletedScan{}, scanNotFinishedError{
			ScanID:   strconv.Itoa(resource.ScanID),
			ScanName: resource.ScanName,
			SiteID:   strconv.Itoa(resource.SiteID),
			Status:   resource.Status,
		}
	}

	// ignore scans included by name in the blocklist
	if n.ScanBlocklist.Contains(resource.ScanName) {
		return domain.CompletedScan{}, scanNameInBlocklistError{
			ScanID:    strconv.Itoa(resource.ScanID),
			ScanName:  resource.ScanName,
			SiteID:    strconv.Itoa(resource.SiteID),
			Blocklist: n.ScanBlocklist,
		}
	}

	// extract scan end time from scan resource
	endTime, err := time.Parse(time.RFC3339Nano, resource.EndTime)
	if err != nil {
		return domain.CompletedScan{}, err
	}

	// scans are fetched sorted by end time in descending order,
	// so the first scan resource before or equal to the start
	// time signals that no more scans need to be processed
	if endTime.Before(start) || endTime.Equal(start) {
		return domain.CompletedScan{}, outOfRangeError{
			ScanID:   strconv.Itoa(resource.ScanID),
			ScanName: resource.ScanName,
			SiteID:   strconv.Itoa(resource.SiteID),
			Start:    start,
			ScanTime: endTime,
		}
	}

	// extract scan end time from scan resource
	startTime, err := time.Parse(time.RFC3339Nano, resource.StartTime)
	if err != nil {
		return domain.CompletedScan{}, err
	}

	return domain.CompletedScan{
		SiteID:    strconv.Itoa(resource.SiteID),
		ScanID:    strconv.Itoa(resource.ScanID),
		ScanType:  resource.ScanType,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

// CheckDependencies makes a call to the nexpose endppoint "/api/3".
// Because asset producer endpoints vary user to user, we want to hit an endpoint
// that is consistent for any Nexpose user
func (n *NexposeClient) CheckDependencies(ctx context.Context) error {
	u, _ := url.Parse(n.Endpoint.String())
	u.Path = path.Join("/api/3")
	req, _ := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	res, err := n.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("Nexpose unexpectedly returned non-200 response code: %d attempting to GET: %s. ", res.StatusCode, u.String())
	}

	return nil
}

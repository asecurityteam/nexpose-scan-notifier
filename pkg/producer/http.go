package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
)

// HTTP holds configuration for producing completed scan events to an HTTP endpoint
type HTTP struct {
	Client   *http.Client
	Endpoint *url.URL
}

type scanPayload struct {
	ScanID string `json:"scanID,omitempty"`
	SiteID string `json:"siteID,omitempty"`
}

// Produce sends the completed scan event to an HTTP endpoint
func (p *HTTP) Produce(ctx context.Context, scan domain.CompletedScan) error {
	payload := scanPayload{
		ScanID: scan.ScanID,
		SiteID: scan.SiteID,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, p.Endpoint.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := p.Client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from http producer: %d %s",
			res.StatusCode, string(resBody))
	}
	return nil
}

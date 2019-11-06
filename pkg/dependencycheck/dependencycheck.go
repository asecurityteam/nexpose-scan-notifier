package dependencycheck

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
)

// DependencyCheck implements the interfaces to fetch vulnerabilities and solutions from nexpose
type DependencyCheck struct {
	HTTPClient       *http.Client
	NexposeEndpoint  *url.URL
	TimestampFetcher domain.TimestampFetcher
}

// CheckDependencies fetches the solutions to a particular vulnerability
func (dc *DependencyCheck) CheckDependencies(ctx context.Context) error {
	u, _ := url.Parse(dc.NexposeEndpoint.String())
	req, _ := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	res, err := dc.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("Nexpose unexpectedly returned non-200 response code: %d attempting to GET: %s", res.StatusCode, u.String())
	}

	_, err = dc.TimestampFetcher.FetchTimestamp(ctx)
	switch err.(type) {
	case nil:
	case domain.TimestampNotFound:
	default:
		return err
	}
	return nil
}

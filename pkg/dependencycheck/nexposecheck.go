package dependencycheck

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// DependencyCheck implements the interfaces to fetch vulnerabilities and solutions from nexpose
type DependencyCheck struct {
	HTTPClient *http.Client
	Endpoint   *url.URL
}

// DepCheck fetches the solutions to a particular vulnerability
func (dc *DependencyCheck) DepCheck(ctx context.Context) error {
	u, _ := url.Parse(dc.Endpoint.String() + "/api/3/scans")
	req, _ := http.NewRequest(http.MethodGet, u.String(), http.NoBody)
	res, err := dc.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		// Nexpose is down?  Nexpose unexpectedly gave 404?

		return fmt.Errorf("Nexpose unexpectedly returned non-200 response code: %d attempting to GET: %s", res.StatusCode, u.String())
	}

	return nil
}

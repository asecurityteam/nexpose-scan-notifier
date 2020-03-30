package scanfetcher

import (
	"fmt"
	"time"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/container"
)

// outOfRangeError is an error indicating the scan time was outside the valid range.
type outOfRangeError struct {
	ScanID   string
	ScanName string
	SiteID   string
	Start    time.Time
	ScanTime time.Time
}

func (e outOfRangeError) Error() string {
	return fmt.Sprintf("scan %s (\"%s\") for site %s was before start time %s: %s",
		e.ScanID, e.ScanName, e.SiteID, e.Start.Format(time.RFC3339Nano), e.ScanTime.Format(time.RFC3339Nano))
}

// scanNotFinishedError is an error indicating the scan status is not "finished".
type scanNotFinishedError struct {
	ScanID   string
	ScanName string
	SiteID   string
	Status   string
}

func (e scanNotFinishedError) Error() string {
	return fmt.Sprintf("scan %s (\"%s\") for site %s status %s does not match %s",
		e.ScanID, e.ScanName, e.SiteID, e.Status, finishedScanStatus)
}

// scanNameInBlocklistError in an error indicating the scan's name is in the
// service's block list, and should therefore be discarded.
type scanNameInBlocklistError struct {
	ScanID    string
	ScanName  string
	SiteID    string
	Blocklist *container.StringContainer
}

func (e scanNameInBlocklistError) Error() string {
	return fmt.Sprintf("scan %s (\"%s\") for site %s in blocklist: %s",
		e.ScanID, e.ScanName, e.SiteID, e.Blocklist)
}

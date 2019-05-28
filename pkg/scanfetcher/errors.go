package scanfetcher

import (
	"fmt"
	"time"
)

// outOfRangeError is an error indicating the scan time was outside the valid range.
type outOfRangeError struct {
	ScanID   string
	SiteID   string
	Start    time.Time
	ScanTime time.Time
}

func (e outOfRangeError) Error() string {
	return fmt.Sprintf("scan %s for site %s was before start time %s: %s",
		e.ScanID, e.SiteID, e.Start.Format(time.RFC3339Nano), e.ScanTime.Format(time.RFC3339Nano))
}

// scanNotFinishedError is an error indicating the scan status is not "finished".
type scanNotFinishedError struct {
	ScanID string
	SiteID string
	Status string
}

func (e scanNotFinishedError) Error() string {
	return fmt.Sprintf("scan %s for site %s status %s does not match %s",
		e.ScanID, e.SiteID, e.Status, finishedScanStatus)
}

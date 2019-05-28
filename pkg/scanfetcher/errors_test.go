package scanfetcher

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOutOfRangeError(t *testing.T) {
	now := time.Now()
	e := outOfRangeError{ScanID: "1", SiteID: "1", Start: now, ScanTime: now.Add(1 * time.Minute)}
	require.Equal(t, e.Error(), fmt.Sprintf("scan 1 for site 1 was before start time %s: %s",
		now.Format(time.RFC3339Nano), now.Add(1*time.Minute).Format(time.RFC3339Nano)))
}

func TestScanNotFinishedError(t *testing.T) {
	e := scanNotFinishedError{ScanID: "1", SiteID: "1", Status: "running"}
	require.Equal(t, e.Error(), "scan 1 for site 1 status running does not match finished")
}

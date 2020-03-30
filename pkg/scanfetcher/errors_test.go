package scanfetcher

import (
	"fmt"
	"testing"
	"time"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/container"
	"github.com/stretchr/testify/require"
)

func TestOutOfRangeError(t *testing.T) {
	now := time.Now()
	e := outOfRangeError{ScanID: "1", ScanName: "Test", SiteID: "1", Start: now, ScanTime: now.Add(1 * time.Minute)}
	require.Equal(t, e.Error(), fmt.Sprintf("scan 1 (\"Test\") for site 1 was before start time %s: %s",
		now.Format(time.RFC3339Nano), now.Add(1*time.Minute).Format(time.RFC3339Nano)))
}

func TestScanNotFinishedError(t *testing.T) {
	e := scanNotFinishedError{ScanID: "1", ScanName: "Test", SiteID: "1", Status: "running"}
	require.Equal(t, e.Error(), "scan 1 (\"Test\") for site 1 status running does not match finished")
}

func TestScanNameInBlocklistError(t *testing.T) {
	e := scanNameInBlocklistError{ScanID: "1", ScanName: "Test", SiteID: "1", Blocklist: container.NewStringContainer([]string{"Bad Scan"})}
	require.Equal(t, e.Error(), "scan 1 (\"Test\") for site 1 in blocklist: [Bad Scan]")
}

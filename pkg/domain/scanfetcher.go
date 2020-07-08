package domain

import (
	"context"
	"time"
)

// CompletedScan represents identifiers for a completed Nexpose scan.
type CompletedScan struct {
	ScanID    string
	SiteID    string
	ScanType  string
	StartTime time.Time
	EndTime   time.Time
}

// ScanFetcher fetchs scans completed from the provided time until now.
type ScanFetcher interface {
	FetchScans(context.Context, time.Time) ([]CompletedScan, error)
}

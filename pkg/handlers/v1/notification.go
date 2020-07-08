package v1

import (
	"context"
	"sort"
	"time"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
	"github.com/asecurityteam/nexpose-scan-notifier/pkg/logs"
)

// Output contains a list of completed Nexpose scans.
type Output struct {
	Response []scanNotification `json:"response"`
}

// scanNotification represents a completed scan event.
type scanNotification struct {
	ScanID    string `json:"scanID"`
	SiteID    string `json:"siteID"`
	ScanType  string `json:"scanType"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

// NotificationHandler takes a duration and returns a list of completed scans.
type NotificationHandler struct {
	ScanFetcher      domain.ScanFetcher
	TimestampFetcher domain.TimestampFetcher
	TimestampStorer  domain.TimestampStorer
	Producer         domain.Producer
	LogFn            domain.LogFn
	StatFn           domain.StatFn
}

// Handle queries for completed scans since the last known successfully processed
// scan timestamp, produces all completed scans to a queue, and returns the list
// of completed scans.
func (h *NotificationHandler) Handle(ctx context.Context) (Output, error) {
	logger := h.LogFn(ctx)
	stater := h.StatFn(ctx)

	lastScanTimestamp, err := h.TimestampFetcher.FetchTimestamp(ctx)
	switch err.(type) {
	case nil:
	case domain.TimestampNotFound:
	default:
		logger.Error(logs.StorageFailure{Reason: err.Error()})
		return Output{}, err
	}

	scans, err := h.ScanFetcher.FetchScans(ctx, lastScanTimestamp)
	if err != nil {
		logger.Error(logs.ScanFetcherFailure{Reason: err.Error()})
		return Output{}, err
	}

	// sort scans by earliest time completed
	sort.SliceStable(scans, func(left, right int) bool {
		return scans[left].EndTime.Before(scans[right].EndTime)
	})

	scanNotifications := make([]scanNotification, len(scans))
	for offset, scan := range scans {
		// Produce completed scan events to a queue
		err := h.Producer.Produce(ctx, scan)
		if err != nil {
			logger.Error(logs.ProducerFailure{Reason: err.Error()})
			return Output{}, err
		}
		if err := h.TimestampStorer.StoreTimestamp(ctx, scan.EndTime); err != nil {
			return Output{}, err
		}
		// emit a statistic of the time between a completed scan and the scan is produced
		stater.Timing("scannotificationdelay", time.Since(scan.EndTime))
		scanNotifications[offset] = completedScanToScanNotification(scan)
	}
	return Output{Response: scanNotifications}, nil
}

func completedScanToScanNotification(scan domain.CompletedScan) scanNotification {
	return scanNotification{
		ScanID:    scan.ScanID,
		SiteID:    scan.SiteID,
		ScanType:  scan.ScanType,
		StartTime: scan.StartTime.Format(time.RFC3339Nano),
		EndTime:   scan.EndTime.Format(time.RFC3339Nano),
	}
}

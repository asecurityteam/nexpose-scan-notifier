package v1

import (
	"context"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
	"github.com/asecurityteam/nexpose-scan-notifier/pkg/logs"
)

// Output contains a list of completed Nexpose scans.
type Output struct {
	Response []scanNotification `json:"response"`
}

// scanNotification represents
type scanNotification struct {
	ScanID string `json:"scanID"`
	SiteID string `json:"siteID"`
}

// NotificationHandler takes a duration and returns a list of completed scans.
type NotificationHandler struct {
	ScanFetcher      domain.ScanFetcher
	TimestampFetcher domain.TimestampFetcher
	TimestampStorer  domain.TimestampStorer
	Producer         domain.Producer
	LogFn            domain.LogFn
}

// Handle accepts takes Input with a duration, queries for completed scans within the duration,
// produces all completed scans to a queue, and returns the list of completed scans.
func (h *NotificationHandler) Handle(ctx context.Context) (Output, error) {
	logger := h.LogFn(ctx)

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

	scanNotifications := make([]scanNotification, len(scans))
	for i, scan := range scans {
		// Produce completed scans to a queue
		err := h.Producer.Produce(ctx, scan)
		if err != nil {
			logger.Error(logs.ProducerFailure{Reason: err.Error()})
			return Output{}, err
		}
		if err := h.TimestampStorer.StoreTimestamp(ctx, scan.Timestamp); err != nil {
			return Output{}, err
		}
		scanNotifications[i] = completedScanToScanNotification(scan)
	}
	return Output{Response: scanNotifications}, nil
}

func completedScanToScanNotification(scan domain.CompletedScan) scanNotification {
	return scanNotification{
		ScanID: scan.ScanID,
		SiteID: scan.ScanID,
	}
}

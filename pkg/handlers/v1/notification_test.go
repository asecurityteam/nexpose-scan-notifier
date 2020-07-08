package v1

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
	"github.com/stretchr/testify/require"
)

func TestCompletedScanToscanNotification(t *testing.T) {
	scanID := "1"
	siteID := "1"
	now := time.Now()
	scan := completedScanToScanNotification(domain.CompletedScan{
		SiteID:    siteID,
		ScanID:    scanID,
		EndTime:   now,
		StartTime: now.Add(time.Second * -10),
	})
	require.Equal(t, scanID, scan.ScanID)
	require.Equal(t, siteID, scan.SiteID)
}

func TestHandle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ts := time.Now().Add(-1 * time.Hour)
	tc := []struct {
		Name               string
		Timestamp          time.Time
		FetchTimestampErr  error
		ExpectFetchScan    bool
		Scans              []domain.CompletedScan
		FetchScanErr       error
		ProducerErrs       []error
		StoreTimestampErrs []error
		Output             Output
		Err                error
	}{
		{
			Name:              "success single scan",
			Timestamp:         ts,
			FetchTimestampErr: nil,
			ExpectFetchScan:   true,
			Scans: []domain.CompletedScan{
				{
					ScanID:    "1",
					SiteID:    "11",
					ScanType:  "Scheduled",
					StartTime: ts.Add(time.Second),
					EndTime:   ts.Add(time.Second * 10),
				},
			},
			FetchScanErr:       nil,
			ProducerErrs:       []error{nil},
			StoreTimestampErrs: []error{nil},
			Output: Output{
				Response: []scanNotification{
					{
						ScanID:    "1",
						SiteID:    "11",
						ScanType:  "Scheduled",
						StartTime: ts.Add(time.Second).Format(time.RFC3339Nano),
						EndTime:   ts.Add(time.Second * 10).Format(time.RFC3339Nano),
					},
				},
			},
			Err: nil,
		},
		{
			Name:              "success with no timestamp found",
			Timestamp:         time.Time{},
			FetchTimestampErr: domain.TimestampNotFound{},
			ExpectFetchScan:   true,
			Scans: []domain.CompletedScan{
				{
					ScanID:    "1",
					SiteID:    "11",
					ScanType:  "Scheduled",
					StartTime: ts.Add(time.Second),
					EndTime:   ts.Add(time.Second * 10),
				},
			},
			FetchScanErr:       nil,
			ProducerErrs:       []error{nil},
			StoreTimestampErrs: []error{nil},
			Output: Output{
				Response: []scanNotification{
					{
						ScanID:    "1",
						SiteID:    "11",
						ScanType:  "Scheduled",
						StartTime: ts.Add(time.Second).Format(time.RFC3339Nano),
						EndTime:   ts.Add(time.Second * 10).Format(time.RFC3339Nano),
					},
				},
			},
			Err: nil,
		},
		{
			Name:               "timestamp fetch error",
			Timestamp:          time.Time{},
			FetchTimestampErr:  fmt.Errorf("timestamp fetch error"),
			ExpectFetchScan:    false,
			Scans:              nil,
			FetchScanErr:       nil,
			ProducerErrs:       nil,
			StoreTimestampErrs: nil,
			Output:             Output{},
			Err:                fmt.Errorf("timestamp fetch error"),
		},
		{
			Name:               "fetch scan error",
			Timestamp:          ts,
			FetchTimestampErr:  nil,
			ExpectFetchScan:    true,
			Scans:              nil,
			FetchScanErr:       fmt.Errorf("fetch scan error"),
			ProducerErrs:       nil,
			StoreTimestampErrs: nil,
			Output:             Output{},
			Err:                fmt.Errorf("fetch scan error"),
		},
		{
			Name:              "producer error",
			Timestamp:         ts,
			FetchTimestampErr: nil,
			ExpectFetchScan:   true,
			Scans: []domain.CompletedScan{
				{
					ScanID:    "1",
					SiteID:    "11",
					StartTime: ts.Add(1 * time.Second),
					EndTime:   ts.Add(2 * time.Second),
				},
				{
					ScanID:    "2",
					SiteID:    "22",
					StartTime: ts.Add(3 * time.Second),
					EndTime:   ts.Add(4 * time.Second),
				},
			},
			FetchScanErr:       nil,
			ProducerErrs:       []error{nil, fmt.Errorf("producer error")},
			StoreTimestampErrs: []error{nil},
			Output:             Output{},
			Err:                fmt.Errorf("producer error"),
		},
		{
			Name:              "store timestamp error",
			Timestamp:         ts,
			FetchTimestampErr: nil,
			ExpectFetchScan:   true,
			Scans: []domain.CompletedScan{
				{
					ScanID:    "1",
					SiteID:    "11",
					StartTime: ts.Add(1 * time.Second),
					EndTime:   ts.Add(2 * time.Second),
				},
			},
			FetchScanErr:       nil,
			ProducerErrs:       []error{nil},
			StoreTimestampErrs: []error{fmt.Errorf("store timestamp error")},
			Output:             Output{},
			Err:                fmt.Errorf("store timestamp error"),
		},
	}

	mockScanFetcher := NewMockScanFetcher(ctrl)
	mockTimestampFetcher := NewMockTimestampFetcher(ctrl)
	mockTimestampStorer := NewMockTimestampStorer(ctrl)
	mockProducer := NewMockProducer(ctrl)

	handler := NotificationHandler{
		LogFn:            testLogFn,
		ScanFetcher:      mockScanFetcher,
		TimestampFetcher: mockTimestampFetcher,
		TimestampStorer:  mockTimestampStorer,
		Producer:         mockProducer,
		StatFn:           MockStatFn,
	}

	for _, tt := range tc {
		t.Run(tt.Name, func(t *testing.T) {
			mockTimestampFetcher.EXPECT().FetchTimestamp(gomock.Any()).Return(tt.Timestamp, tt.FetchTimestampErr)
			if tt.ExpectFetchScan {
				mockScanFetcher.EXPECT().FetchScans(gomock.Any(), tt.Timestamp).Return(tt.Scans, tt.FetchScanErr)
			}
			for _, err := range tt.ProducerErrs {
				mockProducer.EXPECT().Produce(gomock.Any(), gomock.Any()).Return(err)
			}
			for _, err := range tt.StoreTimestampErrs {
				mockTimestampStorer.EXPECT().StoreTimestamp(gomock.Any(), gomock.Any()).Return(err)
			}

			output, err := handler.Handle(context.Background())
			require.Equal(t, tt.Output, output)
			require.Equal(t, tt.Err, err)
		})
	}
}

func TestHandleSortOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ts := time.Now().Add(-1 * time.Hour)
	tc := []struct {
		Name      string
		Timestamp time.Time
		Scans     []domain.CompletedScan
		Output    Output
	}{
		{
			Name:      "scans fetched in order",
			Timestamp: ts,
			Scans: []domain.CompletedScan{
				{
					ScanID:    "1",
					SiteID:    "11",
					ScanType:  "Scheduled",
					StartTime: ts,
					EndTime:   ts.Add(1 * time.Second),
				},
				{
					ScanID:    "2",
					SiteID:    "22",
					ScanType:  "Scheduled",
					StartTime: ts,
					EndTime:   ts.Add(2 * time.Second),
				},
			},
			Output: Output{
				Response: []scanNotification{
					{
						ScanID:    "1",
						SiteID:    "11",
						ScanType:  "Scheduled",
						StartTime: ts.Format(time.RFC3339Nano),
						EndTime:   ts.Add(1 * time.Second).Format(time.RFC3339Nano),
					},
					{
						ScanID:    "2",
						SiteID:    "22",
						ScanType:  "Scheduled",
						StartTime: ts.Format(time.RFC3339Nano),
						EndTime:   ts.Add(2 * time.Second).Format(time.RFC3339Nano),
					},
				},
			},
		},
		{
			Name:      "scans fetched out of order",
			Timestamp: ts,
			Scans: []domain.CompletedScan{
				{
					ScanID:    "4",
					SiteID:    "44",
					ScanType:  "Scheduled",
					StartTime: ts,
					EndTime:   ts.Add(4 * time.Second),
				},
				{
					ScanID:    "1",
					SiteID:    "11",
					ScanType:  "Scheduled",
					StartTime: ts,
					EndTime:   ts.Add(1 * time.Second),
				},
				{
					ScanID:    "3",
					SiteID:    "33",
					ScanType:  "Scheduled",
					StartTime: ts,
					EndTime:   ts.Add(3 * time.Second),
				},
				{
					ScanID:    "2",
					SiteID:    "22",
					ScanType:  "Scheduled",
					StartTime: ts,
					EndTime:   ts.Add(2 * time.Second),
				},
			},
			Output: Output{
				Response: []scanNotification{
					{
						ScanID:    "1",
						SiteID:    "11",
						ScanType:  "Scheduled",
						StartTime: ts.Format(time.RFC3339Nano),
						EndTime:   ts.Add(1 * time.Second).Format(time.RFC3339Nano),
					},
					{
						ScanID:    "2",
						SiteID:    "22",
						ScanType:  "Scheduled",
						StartTime: ts.Format(time.RFC3339Nano),
						EndTime:   ts.Add(2 * time.Second).Format(time.RFC3339Nano),
					},
					{
						ScanID:    "3",
						SiteID:    "33",
						ScanType:  "Scheduled",
						StartTime: ts.Format(time.RFC3339Nano),
						EndTime:   ts.Add(3 * time.Second).Format(time.RFC3339Nano),
					},
					{
						ScanID:    "4",
						SiteID:    "44",
						ScanType:  "Scheduled",
						StartTime: ts.Format(time.RFC3339Nano),
						EndTime:   ts.Add(4 * time.Second).Format(time.RFC3339Nano),
					},
				},
			},
		},
		{
			Name:      "scans fetched out of order with duplicate timestamps",
			Timestamp: ts,
			Scans: []domain.CompletedScan{
				{
					ScanID:    "4",
					SiteID:    "44",
					ScanType:  "Scheduled",
					StartTime: ts,
					EndTime:   ts.Add(3 * time.Second),
				},
				{
					ScanID:    "1",
					SiteID:    "11",
					ScanType:  "Scheduled",
					StartTime: ts,
					EndTime:   ts.Add(1 * time.Second),
				},
				{
					ScanID:    "3",
					SiteID:    "33",
					ScanType:  "Scheduled",
					StartTime: ts,
					EndTime:   ts.Add(3 * time.Second),
				},
				{
					ScanID:    "2",
					SiteID:    "22",
					ScanType:  "Scheduled",
					StartTime: ts,
					EndTime:   ts.Add(2 * time.Second),
				},
			},
			Output: Output{
				Response: []scanNotification{
					{
						ScanID:    "1",
						SiteID:    "11",
						ScanType:  "Scheduled",
						StartTime: ts.Format(time.RFC3339Nano),
						EndTime:   ts.Add(1 * time.Second).Format(time.RFC3339Nano),
					},
					{
						ScanID:    "2",
						SiteID:    "22",
						ScanType:  "Scheduled",
						StartTime: ts.Format(time.RFC3339Nano),
						EndTime:   ts.Add(2 * time.Second).Format(time.RFC3339Nano),
					},
					{
						ScanID:    "4",
						SiteID:    "44",
						ScanType:  "Scheduled",
						StartTime: ts.Format(time.RFC3339Nano),
						EndTime:   ts.Add(3 * time.Second).Format(time.RFC3339Nano),
					},
					{
						ScanID:    "3",
						SiteID:    "33",
						ScanType:  "Scheduled",
						StartTime: ts.Format(time.RFC3339Nano),
						EndTime:   ts.Add(3 * time.Second).Format(time.RFC3339Nano),
					},
				},
			},
		},
	}

	mockScanFetcher := NewMockScanFetcher(ctrl)
	mockTimestampFetcher := NewMockTimestampFetcher(ctrl)
	mockTimestampStorer := NewMockTimestampStorer(ctrl)
	mockProducer := NewMockProducer(ctrl)

	handler := NotificationHandler{
		LogFn:            testLogFn,
		ScanFetcher:      mockScanFetcher,
		TimestampFetcher: mockTimestampFetcher,
		TimestampStorer:  mockTimestampStorer,
		Producer:         mockProducer,
		StatFn:           MockStatFn,
	}

	for _, tt := range tc {
		t.Run(tt.Name, func(t *testing.T) {
			mockTimestampFetcher.EXPECT().FetchTimestamp(gomock.Any()).Return(tt.Timestamp, nil)
			mockScanFetcher.EXPECT().FetchScans(gomock.Any(), tt.Timestamp).Return(tt.Scans, nil)
			for range tt.Scans {
				mockProducer.EXPECT().Produce(gomock.Any(), gomock.Any()).Return(nil)
			}
			for range tt.Scans {
				mockTimestampStorer.EXPECT().StoreTimestamp(gomock.Any(), gomock.Any()).Return(nil)
			}

			output, _ := handler.Handle(context.Background())
			require.Equal(t, tt.Output, output)
		})
	}
}

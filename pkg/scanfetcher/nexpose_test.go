package scanfetcher

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	http "net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNexposeClient_FetchScans(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	endpoint, _ := url.Parse("http://localhost")
	timestamp := time.Date(2019, 05, 24, 00, 00, 00, 00, time.UTC)
	afterTimestamp := time.Date(2019, 05, 25, 00, 00, 00, 00, time.UTC)
	beforeTimestamp := time.Date(2019, 05, 23, 00, 00, 00, 00, time.UTC)
	invalidTimestamp := "2019-05-28 1PM"
	testScanResponse := `
		{
			"resources": [
				{
					"endTime": "%s",
					"id": 1001,
					"siteId": 1,
					"status": "%s"
				}
			],
			"page": {
				"number": %d,
				"size": 1,
				"totalResources": 3,
				"totalPages": %d
			}
		}`

	tests := []struct {
		name         string
		responses    []*http.Response
		responseErrs []error
		expected     []domain.CompletedScan
		expectErr    bool
	}{
		{
			name: "success",
			responses: []*http.Response{
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(testScanResponse,
						afterTimestamp.Format(time.RFC3339Nano), finishedScanStatus, 0, 1)))),
				},
			},
			responseErrs: []error{nil},
			expected: []domain.CompletedScan{
				{
					Timestamp: afterTimestamp,
					ScanID:    "1001",
					SiteID:    "1",
				},
			},
			expectErr: false,
		},
		{
			name: "success with running scan, one scan after timestamp, one scan before timestamp",
			responses: []*http.Response{
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(testScanResponse,
						afterTimestamp.Format(time.RFC3339Nano), "running", 0, 3)))),
				},
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(testScanResponse,
						afterTimestamp.Format(time.RFC3339Nano), finishedScanStatus, 1, 3)))),
				},
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(testScanResponse,
						beforeTimestamp.Format(time.RFC3339Nano), finishedScanStatus, 2, 3)))),
				},
			},
			responseErrs: []error{nil, nil, nil},
			expected: []domain.CompletedScan{
				{
					Timestamp: afterTimestamp,
					ScanID:    "1001",
					SiteID:    "1",
				},
			},
			expectErr: false,
		},
		{
			name: "out of range error for one scan",
			responses: []*http.Response{
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(testScanResponse,
						beforeTimestamp.Format(time.RFC3339Nano), finishedScanStatus, 0, 1)))),
				},
			},
			responseErrs: []error{nil},
			expected:     nil,
			expectErr:    false,
		},
		{
			name: "nexpose error",
			responses: []*http.Response{
				nil,
			},
			responseErrs: []error{fmt.Errorf("request error")},
			expected:     nil,
			expectErr:    true,
		},
		{
			name: "error fetching another page of scans",
			responses: []*http.Response{
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(testScanResponse,
						afterTimestamp.Format(time.RFC3339Nano), finishedScanStatus, 0, 2)))),
				},
				nil,
			},
			responseErrs: []error{nil, fmt.Errorf("nexpose error")},
			expected:     nil,
			expectErr:    true,
		},
		{
			name: "one page with invalid timestamp",
			responses: []*http.Response{
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(testScanResponse,
						invalidTimestamp, finishedScanStatus, 0, 1)))),
				},
			},
			responseErrs: []error{nil},
			expected:     nil,
			expectErr:    true,
		},
		{
			name: "success with first page, error on second page with scan with invalid timestamp",
			responses: []*http.Response{
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(testScanResponse,
						afterTimestamp.Format(time.RFC3339Nano), finishedScanStatus, 0, 2)))),
				},
				&http.Response{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf(testScanResponse,
						invalidTimestamp, finishedScanStatus, 1, 2)))),
				},
			},
			responseErrs: []error{nil, nil},
			expected:     nil,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRT := NewMockRoundTripper(ctrl)
			for offset := range tt.responses {
				mockRT.EXPECT().RoundTrip(gomock.Any()).Return(tt.responses[offset], tt.responseErrs[offset])
			}
			nexposeClient := &NexposeClient{
				Client:   &http.Client{Transport: mockRT},
				Endpoint: endpoint,
			}
			actual, err := nexposeClient.FetchScans(context.Background(), timestamp)
			require.Equal(t, tt.expected, actual)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.Nil(t, err)
		})
	}
}

type errReader struct {
	Error error
}

func (r *errReader) Read(_ []byte) (int, error) {
	return 0, r.Error
}

func TestNexposeClient_makePagedNexposeScanRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	endpoint, _ := url.Parse("http://localhost")
	ts := time.Date(2019, 05, 24, 00, 01, 00, 00, time.UTC)
	testScanResponse := fmt.Sprintf(`
		{
			"resources": [
				{
					"endTime": "%s",
					"id": 1001,
					"siteId": 1,
					"status": "running"
				}
			],
			"page": {
				"number": 1,
				"size": 1,
				"totalResources": 200,
				"totalPages": 200
			}
		}`, ts.Format(time.RFC3339Nano))

	tests := []struct {
		name        string
		response    *http.Response
		responseErr error
		expected    nexposeScanResponse
		expectErr   bool
	}{
		{
			name:        "success",
			response:    &http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(testScanResponse)))},
			responseErr: nil,
			expected: nexposeScanResponse{
				Resources: []resource{
					{
						EndTime: ts.Format(time.RFC3339Nano),
						ScanID:  1001,
						SiteID:  1,
						Status:  "running",
					},
				},
				Page: page{
					Number:         1,
					Size:           1,
					TotalPages:     200,
					TotalResources: 200,
				},
			},
			expectErr: false,
		},
		{
			name:        "response error",
			response:    nil,
			responseErr: fmt.Errorf("response error"),
			expected:    nexposeScanResponse{},
			expectErr:   true,
		},
		{
			name:        "io read error",
			response:    &http.Response{Body: ioutil.NopCloser(&errReader{Error: fmt.Errorf("io read error")})},
			responseErr: nil,
			expected:    nexposeScanResponse{},
			expectErr:   true,
		},
		{
			name:        "invalid json error",
			response:    &http.Response{Body: ioutil.NopCloser(bytes.NewBuffer([]byte(`{notjson}`)))},
			responseErr: nil,
			expected:    nexposeScanResponse{},
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRT := NewMockRoundTripper(ctrl)
			mockRT.EXPECT().RoundTrip(gomock.Any()).Return(tt.response, tt.responseErr)
			nexposeClient := &NexposeClient{
				Client:   &http.Client{Transport: mockRT},
				Endpoint: endpoint,
			}
			actual, err := nexposeClient.makePagedNexposeScanRequest(0)
			require.Equal(t, tt.expected, actual)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.Nil(t, err)
		})
	}
}

func TestScanResourceToCompletedScan(t *testing.T) {
	start := time.Date(2019, 05, 24, 00, 00, 00, 00, time.UTC)
	beforeStart := time.Date(2019, 05, 23, 00, 00, 00, 00, time.UTC)
	afterStart := time.Date(2019, 05, 25, 00, 00, 00, 00, time.UTC)

	tests := []struct {
		name     string
		resource resource
		expected domain.CompletedScan
		err      error
	}{
		{
			name: "success",
			resource: resource{
				EndTime: afterStart.Format(time.RFC3339Nano),
				ScanID:  1001,
				SiteID:  1,
				Status:  finishedScanStatus,
			},
			expected: domain.CompletedScan{
				SiteID:    strconv.Itoa(1),
				ScanID:    strconv.Itoa(1001),
				Timestamp: afterStart,
			},
			err: nil,
		},
		{
			name: "scan out of range",
			resource: resource{
				EndTime: beforeStart.Format(time.RFC3339Nano),
				ScanID:  1001,
				SiteID:  1,
				Status:  finishedScanStatus,
			},
			expected: domain.CompletedScan{},
			err:      outOfRangeError{},
		},
		{
			name: "scan not finished",
			resource: resource{
				EndTime: afterStart.Format(time.RFC3339Nano),
				ScanID:  1001,
				SiteID:  1,
				Status:  "running",
			},
			expected: domain.CompletedScan{},
			err:      fmt.Errorf("scan not finished"),
		},
		{
			name: "end time not parseable",
			resource: resource{
				EndTime: "",
				ScanID:  1001,
				SiteID:  1,
				Status:  finishedScanStatus,
			},
			expected: domain.CompletedScan{},
			err:      fmt.Errorf("end time not parseable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := scanResourceToCompletedScan(tt.resource, start)
			require.Equal(t, tt.expected, actual)
			if tt.err != nil {
				require.Error(t, err)
				return
			}
			require.Nil(t, err)
		})
	}
}

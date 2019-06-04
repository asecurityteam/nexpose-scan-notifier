package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type errReader struct {
	Error error
}

func (r *errReader) Read(_ []byte) (int, error) {
	return 0, r.Error
}

func TestHTTP_Produce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)

	scan := domain.CompletedScan{
		ScanID: "1",
		SiteID: "2",
	}
	respJSON, _ := json.Marshal(scan)
	respReader := bytes.NewReader(respJSON)
	endpoint, _ := url.Parse("http://localhost")
	producer := &HTTP{
		Client:   &http.Client{Transport: mockRT},
		Endpoint: endpoint,
	}

	tests := []struct {
		name        string
		response    *http.Response
		responseErr error
		expectErr   bool
	}{
		{
			name: "success",
			response: &http.Response{
				Body:       ioutil.NopCloser(respReader),
				StatusCode: http.StatusOK,
			},
			responseErr: nil,
			expectErr:   false,
		},
		{
			name:        "request error",
			response:    nil,
			responseErr: errors.New("HTTPError"),
			expectErr:   true,
		},
		{
			name: "non 200 status code",
			response: &http.Response{
				Body:       ioutil.NopCloser(respReader),
				StatusCode: http.StatusNotFound,
			},
			responseErr: nil,
			expectErr:   true,
		},
		{
			name: "io read error",
			response: &http.Response{
				Body:       ioutil.NopCloser(&errReader{Error: fmt.Errorf("io read error")}),
				StatusCode: http.StatusOK,
			},
			responseErr: nil,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRT.EXPECT().RoundTrip(gomock.Any()).Return(tt.response, tt.responseErr)
			err := producer.Produce(context.Background(), scan)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.Nil(t, err)
		})
	}

}

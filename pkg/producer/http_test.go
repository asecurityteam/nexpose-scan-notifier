package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestProduceSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)

	scan := domain.CompletedScan{
		ScanID: "1",
		SiteID: "2",
	}

	respJSON, _ := json.Marshal(scan)
	respReader := bytes.NewReader(respJSON)
	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
		Body:       ioutil.NopCloser(respReader),
		StatusCode: http.StatusOK,
	}, nil)

	endpoint, _ := url.Parse("http://localhost")
	producer := &HTTP{
		Client:   &http.Client{Transport: mockRT},
		Endpoint: endpoint,
	}
	err := producer.Produce(context.Background(), scan)
	require.Nil(t, err)
}

func TestProduceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)

	mockRT.EXPECT().RoundTrip(gomock.Any()).Return(nil, errors.New("HTTPError"))

	scan := domain.CompletedScan{
		ScanID: "1",
		SiteID: "2",
	}
	endpoint, _ := url.Parse("http://localhost")
	producer := &HTTP{
		Client:   &http.Client{Transport: mockRT},
		Endpoint: endpoint,
	}
	err := producer.Produce(context.Background(), scan)
	require.NotNil(t, err)
}

package scanfetcher

import (
	"context"
	"encoding/csv"
	"net/url"
	"strings"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/container"
)

// NexposeConfig holds configuration to connect to Nexpose
// and make a call to the fetch scans API
type NexposeConfig struct {
	Endpoint      string `description:"The scheme and host of a Nexpose instance."`
	PageSize      int    `description:"The number of scans that should be returned from the Nexpose API at one time."`
	ScanBlocklist string `description:"CSV-formatted list of scan names to discard."`
}

// Name is used by the settings library and will add a "NEXPOSE_"
// prefix to NexposeConfig environment variables
func (c *NexposeConfig) Name() string {
	return "Nexpose"
}

// NexposeComponent satisfies the settings library Component
// API, and may be used by the settings.NewComponent function.
type NexposeComponent struct{}

// Settings can be used to populate default values if there are any
func (*NexposeComponent) Settings() *NexposeConfig {
	return &NexposeConfig{
		PageSize:      100,
		ScanBlocklist: "",
	}
}

// New constructs a NexposeClient from a config.
func (*NexposeComponent) New(_ context.Context, c *NexposeConfig) (*NexposeClient, error) {
	endpoint, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, err
	}

	csvReader := csv.NewReader(strings.NewReader(c.ScanBlocklist))
	scanBlockList, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	return &NexposeClient{
		Endpoint:      endpoint,
		PageSize:      c.PageSize,
		ScanBlocklist: container.NewStringContainer(scanBlockList),
	}, nil
}

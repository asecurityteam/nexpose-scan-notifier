package producer

import (
	"context"
	"net/url"
)

// ProducerConfig holds configuration required to send Nexpose assets
// to a queue via an HTTP Producer
type ProducerConfig struct {
	Endpoint string
}

// Name is used by the settings library and will add a "HTTPPRODUCER"
// prefix to ProducerConfig environment variables
func (c *ProducerConfig) Name() string {
	return "HTTPProducer"
}

// ProducerComponent satisfies the settings library Component
// API, and may be used by the settings.NewComponent function.
type ProducerComponent struct{}

// Settings can be used to populate default values if there are any
func (*ProducerComponent) Settings() *ProducerConfig { return &ProducerConfig{} }

// New constructs a HTTP from a config.
func (*ProducerComponent) New(_ context.Context, c *ProducerConfig) (*HTTP, error) {
	endpoint, err := url.Parse(c.Endpoint)
	if err != nil {
		return nil, err
	}

	return &HTTP{Endpoint: endpoint}, nil
}

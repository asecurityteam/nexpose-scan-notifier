package producer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProducerConfig_Name(t *testing.T) {
	producerComponent := ProducerComponent{}
	producerConfig := producerComponent.Settings()
	require.Equal(t, "HTTPProducer", producerConfig.Name())
}

func TestProducerComponent_New(t *testing.T) {
	producerComponent := ProducerComponent{}
	c := &ProducerConfig{
		Endpoint: "http://localhost",
	}
	producer, err := producerComponent.New(context.Background(), c)

	require.Equal(t, "http://localhost", producer.Endpoint.String())
	require.Nil(t, err)
}

func TestProducerComponent_New_InvalidHost(t *testing.T) {
	producerComponent := ProducerComponent{}
	c := &ProducerConfig{
		Endpoint: "~!@#$%^&*()_+:?><!@#$%^&*())_:",
	}
	_, err := producerComponent.New(context.Background(), c)
	require.Error(t, err)
}

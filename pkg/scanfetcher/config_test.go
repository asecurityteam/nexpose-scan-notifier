package scanfetcher

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {
	nexposeClientConfig := NexposeConfig{}
	require.Equal(t, "Nexpose", nexposeClientConfig.Name())
}

func TestComponentDefaultConfig(t *testing.T) {
	component := &NexposeComponent{}
	config := component.Settings()
	require.Empty(t, config.Endpoint)
	require.Equal(t, config.PageSize, 100)
}

func TestNexposeClientConfigWithValues(t *testing.T) {
	nexposeComponent := NexposeComponent{}
	config := &NexposeConfig{
		Endpoint: "http://localhost",
		PageSize: 5,
	}
	nexposeClient, err := nexposeComponent.New(context.Background(), config)

	require.Equal(t, "http://localhost", nexposeClient.Endpoint.String())
	require.Equal(t, 5, nexposeClient.PageSize)
	require.Nil(t, err)
}

func TestNexposeClientConfigWithInvalidEndpoint(t *testing.T) {
	nexposeComponent := NexposeComponent{}
	config := &NexposeConfig{Endpoint: "~!@#$%^&*()_+:?><!@#$%^&*())_:"}
	_, err := nexposeComponent.New(context.Background(), config)

	require.Error(t, err)
}

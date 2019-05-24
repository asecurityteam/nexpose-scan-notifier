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
	component := &NexposeConfigComponent{}
	config := component.Settings()
	require.Empty(t, config.Host)
	require.Equal(t, config.PageSize, 100)
}

func TestNexposeClientConfigWithValues(t *testing.T) {
	nexposeClientComponent := NexposeConfigComponent{}
	config := &NexposeConfig{
		Host:     "http://localhost",
		PageSize: 5,
	}
	nexposeClient, err := nexposeClientComponent.New(context.Background(), config)

	require.Equal(t, "http://localhost", nexposeClient.Host.String())
	require.Equal(t, 5, nexposeClient.PageSize)
	require.Nil(t, err)
}

func TestNexposeClientConfigWithInvalidHost(t *testing.T) {
	nexposeClientComponent := NexposeConfigComponent{}
	config := &NexposeConfig{Host: "~!@#$%^&*()_+:?><!@#$%^&*())_:"}
	_, err := nexposeClientComponent.New(context.Background(), config)

	require.Error(t, err)
}

package scanfetcher

import (
	"context"
	"testing"

	"github.com/asecurityteam/nexpose-scan-notifier/pkg/container"
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
	require.Equal(t, config.ScanBlocklist, "")
}

func TestNexposeClientConfigWithValues(t *testing.T) {
	nexposeComponent := NexposeComponent{}
	config := &NexposeConfig{
		Endpoint:      "http://localhost",
		PageSize:      5,
		ScanBlocklist: "BadScan1,\"Bad Scan, the Second\"",
	}
	nexposeClient, err := nexposeComponent.New(context.Background(), config)

	require.Equal(t, "http://localhost", nexposeClient.Endpoint.String())
	require.Equal(t, 5, nexposeClient.PageSize)
	require.Equal(t, &container.StringContainer{
		"Bad Scan, the Second": struct{}{},
		"BadScan1":             struct{}{},
	}, nexposeClient.ScanBlocklist)
	require.Nil(t, err)
}

func TestNexposeClientConfigWithInvalidEndpoint(t *testing.T) {
	nexposeComponent := NexposeComponent{}
	config := &NexposeConfig{Endpoint: "~!@#$%^&*()_+:?><!@#$%^&*())_:", ScanBlocklist: ""}
	_, err := nexposeComponent.New(context.Background(), config)

	require.Error(t, err)
}

// Note: The [CSV spec](https://tools.ietf.org/html/rfc4180) is vague on this, but Go will
// throw a csv.ParseError for quoted fields with leading spaces, so this is invalid.
func TestNexposeClientConfigWithInvalidBlocklist(t *testing.T) {
	nexposeComponent := NexposeComponent{}
	config := &NexposeConfig{Endpoint: "http://localhost", ScanBlocklist: " \"Bad Scan Name\""}
	_, err := nexposeComponent.New(context.Background(), config)

	require.Error(t, err)
}

package vault

import (
	"net/url"
	"testing"

	"github.com/go-chef/chef"
	"github.com/stretchr/testify/require"
)

func TestVaultURL(t *testing.T) {
	client := &chef.Client{
		BaseURL: &url.URL{Scheme: "https", Host: "localhost"},
	}
	svc := NewService(client)

	got := svc.vaultURL("myvault")
	require.Equal(t, "https://localhost/data/myvault", got)
}

func TestService_getClientsFromSearch_NoQuery(t *testing.T) {
	svc := &Service{}
	payload := &Payload{SearchQuery: nil}

	clients, err := svc.getClientsFromSearch(payload)
	if err != nil {
		t.Fatal(err)
	}

	if len(clients) != 0 {
		t.Fatalf("expected empty client list, got %v", clients)
	}
}

func TestGetClientsFromSearch_WithMatchingQuery(t *testing.T) {
	setupStubs(t)

	query := "name:testhost*"
	clients, err := service.getClientsFromSearch(&Payload{SearchQuery: &query})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, []string{"testhost", "testhost3", "testhost4"}, clients)
}

func TestGetClientsFromSearch_NoQueryReturnsEmpty(t *testing.T) {
	setupStubs(t)

	clients, err := service.getClientsFromSearch(&Payload{})
	require.NoError(t, err)
	require.Empty(t, clients)
}

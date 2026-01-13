package vault

import (
	"reflect"
	"testing"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item_keys"
	"github.com/stretchr/testify/require"
)

func TestRefresh_CleanClients(t *testing.T) {
	setupStubs(t)

	clients := []string{
		"testhost",
		"testhost2",
		"testhost3",
		"testhost4",
		"testhost5",
	}
	kept, removed, err := cleanClients(clients, service.clientExists)
	if err != nil {
		t.Fatal(err)
	}

	wantKept := []string{"testhost", "testhost3", "testhost4"}
	wantRemoved := []string{"testhost2", "testhost5"}
	if !reflect.DeepEqual(kept, wantKept) {
		t.Errorf("kept: %v, wantKept: %v", kept, wantKept)
	}
	if !reflect.DeepEqual(removed, wantRemoved) {
		t.Errorf("removed: %v, wantRemoved: %v", removed, wantRemoved)
	}
}

func TestRefresh_SparseModeEncryptsPerClient(t *testing.T) {
	setupStubs(t)
	var calls []string
	var wrote struct {
		mode item_keys.KeysMode
		keys map[string]any
	}

	deps := refreshOps{
		loadKeysCurrentState: func(*Payload) (*item_keys.VaultItemKeys, error) {
			calls = append(calls, "loadKeys")
			return &item_keys.VaultItemKeys{
				Mode:        item_keys.KeysModeSparse,
				SearchQuery: "name:testhost*",
				Keys: map[string]string{
					"tester": "encrypted secret",
				},
			}, nil
		},
		getClientsFromSearch: func(*Payload) ([]string, error) {
			calls = append(calls, "searchClients")
			return []string{"a", "b"}, nil
		},
		loadSharedSecret: func(*Payload) ([]byte, error) {
			calls = append(calls, "loadSecret")
			return []byte("secret"), nil
		},
		clientPublicKey: func(string) (chef.AccessKey, error) {
			calls = append(calls, "clientKey")
			return chef.AccessKey{
				Name:      "tester",
				PublicKey: "RSA KEY",
			}, nil
		},
		encryptSharedSecret: func(pem string, secret []byte) (string, error) {
			calls = append(calls, "encryptSharedSecret")
			return "encrypted secret", nil
		},
		writeKeys: func(_ *Payload, mode item_keys.KeysMode, keys map[string]any, _ *item_keys.VaultItemKeysResult) error {
			calls = append(calls, "writeKeys")
			wrote.mode = mode
			wrote.keys = keys
			return nil
		},
	}

	_, err := service.refresh(&Payload{}, deps)
	require.NoError(t, err)

	require.Equal(t, []string{
		"loadKeys",
		"searchClients",
		"loadSecret",
		"clientKey",
		"encryptSharedSecret",
		"clientKey",
		"encryptSharedSecret",
		"writeKeys",
	}, calls)
	require.Equal(t, item_keys.KeysModeSparse, wrote.mode)
	keys := wrote.keys
	require.Equal(t, "encrypted secret", keys["tester"])
	require.Equal(t, "encrypted secret", keys["a"])
	require.Equal(t, "encrypted secret", keys["b"])

}

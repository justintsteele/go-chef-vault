package vault

import (
	"reflect"
	"testing"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item_keys"
	"github.com/stretchr/testify/require"
)

type refreshRecorder struct {
	calls []string
	wrote struct {
		keys map[string]any
		mode item_keys.KeysMode
	}
}

func (r *refreshRecorder) ops(keysMode item_keys.KeysMode) refreshOps {
	return refreshOps{
		loadKeysCurrentState: func(*Payload) (*item_keys.VaultItemKeys, error) {
			r.calls = append(r.calls, "loadKeysCurrentState")
			return &item_keys.VaultItemKeys{
				Mode:        keysMode,
				SearchQuery: "name:testhost*",
				Keys: map[string]string{
					"tester": "encrypted secret",
				},
			}, nil
		},
		loadSharedSecret: func(*Payload) ([]byte, error) {
			r.calls = append(r.calls, "loadSecret")
			return []byte("secret"), nil
		},
		clientPublicKey: func(string) (chef.AccessKey, error) {
			r.calls = append(r.calls, "clientKey")
			return chef.AccessKey{
				Name:      "tester",
				PublicKey: "RSA KEY",
			}, nil
		},
		encryptSharedSecret: func(pem string, secret []byte) (string, error) {
			r.calls = append(r.calls, "encryptSharedSecret")
			return "encrypted secret", nil
		},
		writeKeys: func(_ *Payload, mode item_keys.KeysMode, keys map[string]any, _ *item_keys.VaultItemKeysResult) error {
			r.calls = append(r.calls, "writeKeys")
			r.wrote.mode = mode
			r.wrote.keys = keys
			return nil
		},
	}
}

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

	rec := refreshRecorder{}

	_, err := service.refresh(&Payload{}, rec.ops(item_keys.KeysModeSparse))
	require.NoError(t, err)

	require.Equal(t, []string{
		"loadKeysCurrentState",
		"loadSecret",
		"clientKey",
		"encryptSharedSecret",
		"clientKey",
		"encryptSharedSecret",
		"clientKey",
		"encryptSharedSecret",
		"writeKeys",
	}, rec.calls)
	require.Equal(t, item_keys.KeysModeSparse, rec.wrote.mode)
	keys := rec.wrote.keys
	require.Equal(t, "encrypted secret", keys["tester"])
	require.Equal(t, "encrypted secret", keys["testhost3"])
	require.Equal(t, "encrypted secret", keys["testhost4"])
}

func TestRefresh_DefaultKeys(t *testing.T) {
	setupStubs(t)

	rec := refreshRecorder{}

	_, err := service.refresh(&Payload{}, rec.ops(item_keys.KeysModeDefault))
	require.NoError(t, err)

	require.Equal(t, []string{
		"loadKeysCurrentState",
		"loadSecret",
		"clientKey",
		"encryptSharedSecret",
		"clientKey",
		"encryptSharedSecret",
		"clientKey",
		"encryptSharedSecret",
		"writeKeys",
	}, rec.calls)
	require.Equal(t, item_keys.KeysModeDefault, rec.wrote.mode)
	keys := rec.wrote.keys
	require.Equal(t, "encrypted secret", keys["tester"])
	require.Equal(t, "encrypted secret", keys["testhost3"])
	require.Equal(t, "encrypted secret", keys["testhost4"])
}

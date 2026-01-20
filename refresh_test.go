package vault

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item_keys"
	"github.com/stretchr/testify/require"
)

type refreshRecorder struct {
	calls []string
	wrote struct {
		refreshPayload *Payload
		modeState      *item_keys.KeysModeState
	}
}

func (r *refreshRecorder) ops() refreshOps {
	return refreshOps{
		loadSharedSecret: func(*Payload) ([]byte, error) {
			r.calls = append(r.calls, "loadSecret")
			return []byte("secret"), nil
		},
		encryptSharedSecret: func(pem string, secret []byte) (string, error) {
			r.calls = append(r.calls, "encryptSharedSecret")
			return "new encrypted secret", nil
		},
		getItem: func(_, _ string) (chef.DataBagItem, error) {
			r.calls = append(r.calls, "getItem")
			type data chef.DataBagItem
			var current data
			rawCurrent := `{
				"foo": "foo-value-1",
				"bar": "bar-value-1",
				"baz": {
					"fuz": "fuz-value-1",
					"buz": "buz-value-1"
				}
			}`
			if err := json.Unmarshal([]byte(rawCurrent), &current); err != nil {
				return nil, err
			}
			return current, nil
		},
		updateVault: func(payload *Payload, state *item_keys.KeysModeState) (*item_keys.VaultItemKeysResult, error) {
			r.calls = append(r.calls, "updateVault")
			r.wrote.refreshPayload = payload
			r.wrote.modeState = state
			return &item_keys.VaultItemKeysResult{
				URIs: []string{"https://localhost/data/vault1/secret1_keys"},
			}, nil
		},
	}
}

func TestRefresh_CleanUnknownClients(t *testing.T) {
	setupStubs(t)

	clients := []string{
		"testhost",
		"testhost2",
		"testhost3",
		"testhost4",
		"testhost5",
	}
	kept, removed, err := resolveClients(clients, service.clientExists)
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

func TestRefresh_Reencrypt(t *testing.T) {
	setupStubs(t)

	rec := refreshRecorder{}

	_, err := service.refresh(&Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
	}, rec.ops())
	require.NoError(t, err)

	// updateVault kicks off the full workflow to generate a new secret and reencrypt everything
	require.Equal(t, []string{
		"getItem",
		"updateVault",
	}, rec.calls)
}

func TestRefresh_SkipReencrypt(t *testing.T) {
	setupStubs(t)

	rec := refreshRecorder{}

	_, err := service.refresh(&Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		SkipReencrypt: true,
	}, rec.ops())
	require.NoError(t, err)

	// should only encrypt 2 times, new clients only
	require.Equal(t, []string{
		"loadSecret",
		"encryptSharedSecret",
		"encryptSharedSecret",
	}, rec.calls)
}

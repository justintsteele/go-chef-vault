package vault

import (
	"encoding/json"
	"testing"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item_keys"
	"github.com/stretchr/testify/require"
)

type rotateRecorder struct {
	calls []string
	wrote struct {
		rotatePayload *Payload
		modeState     *item_keys.KeysModeState
	}
}

func (r *rotateRecorder) ops() rotateOps {
	return rotateOps{
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
			r.wrote.rotatePayload = payload
			r.wrote.modeState = state
			return &item_keys.VaultItemKeysResult{
				URIs: []string{"https://localhost/data/vault1/secret1_keys"},
			}, nil
		},
	}
}

func TestRotateKeys_DefaultPayload(t *testing.T) {
	setupStubs(t)

	rec := rotateRecorder{}

	_, err := service.rotateKeys(&Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
	}, rec.ops())
	require.NoError(t, err)
	require.Equal(t, []string{"getItem", "updateVault"}, rec.calls)
	require.NotEmpty(t, rec.wrote.rotatePayload.Clients)
	require.NotEmpty(t, rec.wrote.rotatePayload.Admins)
	require.NotEmpty(t, rec.wrote.rotatePayload.Content)
}

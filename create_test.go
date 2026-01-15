package vault

import (
	"encoding/json"
	"testing"

	"github.com/justintsteele/go-chef-vault/item_keys"
	"github.com/stretchr/testify/require"
)

type createRecorder struct {
	calls []string
	wrote struct {
		keysModeState *item_keys.KeysModeState
	}
}

func (r *createRecorder) ops() createOps {
	return createOps{
		createKeysDataBag: func(_ *Payload, keys *item_keys.KeysModeState, secret []byte) (*item_keys.VaultItemKeysResult, error) {
			r.calls = append(r.calls, "createKeysDataBag")
			r.wrote.keysModeState = keys
			return &item_keys.VaultItemKeysResult{
				URIs: []string{"https://localhost/data/vault1/secret1_keys"},
			}, nil
		},
	}
}

func TestCreate_WithDefaultPayload(t *testing.T) {
	setupStubs(t)

	rec := &createRecorder{}

	_, err := service.create(&Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
	}, rec.ops())

	require.NoError(t, err)
	require.Equal(t, item_keys.KeysModeDefault, rec.wrote.keysModeState.Desired)
	require.Equal(t, []string{"createKeysDataBag"}, rec.calls)
}

func TestCreate_WithSparsePayload(t *testing.T) {
	setupStubs(t)

	rec := &createRecorder{}
	var raw map[string]interface{}
	vaultItem := `{"baz": "baz-value-1", "fuz": "fuz-value-2"}`
	if err := json.Unmarshal([]byte(vaultItem), &raw); err != nil {
		panic(err)
	}

	mode := item_keys.KeysModeSparse
	_, err := service.create(&Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		KeysMode:      &mode,
		Content:       raw,
	}, rec.ops())

	require.NoError(t, err)
	require.Equal(t, item_keys.KeysModeSparse, rec.wrote.keysModeState.Desired)
	require.Equal(t, []string{"createKeysDataBag"}, rec.calls)
}

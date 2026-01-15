package vault

import (
	"testing"

	"github.com/justintsteele/go-chef-vault/item_keys"
	"github.com/stretchr/testify/require"
)

type updateRecorder struct {
	calls []string
	wrote struct {
		payload *Payload
		state   *item_keys.KeysModeState
	}
}

func (r *updateRecorder) ops() updateOps {
	return updateOps{
		loadKeysCurrentState: func(*Payload) (*item_keys.VaultItemKeys, error) {
			r.calls = append(r.calls, "loadKeys")
			return &item_keys.VaultItemKeys{
				Mode:        item_keys.KeysModeDefault,
				SearchQuery: "name:testhost*",
				Keys: map[string]string{
					"tester": "encrypted secret",
				},
			}, nil
		},
		resolveUpdateContent: func(p *Payload) (map[string]interface{}, error) {
			r.calls = append(r.calls, "resolveUpdateContent")
			content := map[string]interface{}{
				"foo": "foo-value-1",
				"bar": "bar-value-1",
			}
			return content, nil
		},
		updateVault: func(payload *Payload, state *item_keys.KeysModeState) (*item_keys.VaultItemKeysResult, error) {
			r.calls = append(r.calls, "updateVault")
			r.wrote.payload = payload
			r.wrote.state = state
			return &item_keys.VaultItemKeysResult{
				URIs: []string{"https://localhost/data/vault1/secret1_keys"},
			}, nil
		},
	}
}

func TestUpdate_ChangeKeysMode(t *testing.T) {
	setupStubs(t)

	rec := &updateRecorder{}

	mode := item_keys.KeysModeSparse
	_, err := service.update(&Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		KeysMode:      &mode,
	}, rec.ops())
	require.NoError(t, err)
	require.Equal(t, rec.calls, []string{"loadKeys", "resolveUpdateContent", "updateVault"})
	require.Equal(t, rec.wrote.state.Desired, item_keys.KeysModeSparse)
}

func TestUpdate_NoKeysMode(t *testing.T) {
	setupStubs(t)

	rec := &updateRecorder{}

	_, err := service.update(&Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
	}, rec.ops())
	require.NoError(t, err)
	require.Equal(t, rec.calls, []string{"loadKeys", "resolveUpdateContent", "updateVault"})
	require.Equal(t, rec.wrote.state.Desired, item_keys.KeysModeDefault)
}

func TestUpdate_ResolveUpdateContent(t *testing.T) {
	rawCurrent := map[string]interface{}{
		"foo": "fake-foo-value",
	}
	rawUpdate := map[string]interface{}{
		"bar": "fake-bar-value-2",
	}
	rawMerged := map[string]interface{}{
		"foo": "fake-foo-value",
		"bar": "fake-bar-value-2",
	}

	content, err := resolveContent(rawCurrent, rawUpdate)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, rawMerged, content)
}

func TestPayload_mergeKeyActors_MergesAdminsAndClients(t *testing.T) {
	setupStubs(t)
	payload, err := stubPayload([]string{"tester"}, []string{"testhost"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	keyState, err := service.loadKeysCurrentState(payload)
	if err != nil {
		t.Fatal(err)
	}

	updPayload, err := stubPayload([]string{"tester", "pivotal"}, []string{"testhost", "testhost3"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	updPayload.mergeKeyActors(keyState)

	got := keyState.Admins
	want := []string{"tester", "pivotal"}

	if !item_keys.EqualLists(got, want) {
		t.Errorf("All actors = %v, want %v", got, want)
	}
}

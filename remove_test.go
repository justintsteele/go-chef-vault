package vault

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/go-chef/chef"
	"github.com/stretchr/testify/require"
)

type removeRecorder struct {
	calls []string
	wrote struct {
		resolveContentPayload *Payload
		removePayload         *Payload
	}
}

func (r *removeRecorder) ops() removeOps {
	return removeOps{
		getItem: func(vaultName, vaultItemName string) (chef.DataBagItem, error) {
			r.calls = append(r.calls, "getItem")
			return map[string]interface{}{
				"foo": "foo-value-1",
				"bar": "bar-value-1",
			}, nil
		},
		update: func(payload *Payload) (*UpdateResponse, error) {
			r.calls = append(r.calls, "update")
			r.wrote.removePayload = payload
			return &UpdateResponse{
				Response: Response{
					URI: "https://localhost/data/vault1",
				},
				Data: &UpdateDataResponse{
					URI: "https://localhost/data/vault1/secret1",
				},
				KeysURIs: []string{"https://localhost/data/vault1/secret1_keys", "https://localhost/data/vault1/secret1_key_tester"},
			}, nil
		},
	}
}

func TestRemove_Actor(t *testing.T) {
	setupStubs(t)

	rec := &removeRecorder{}

	_, err := service.remove(&Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		Clients:       []string{"fakehost"},
	}, rec.ops())
	require.NoError(t, err)
	require.Equal(t, []string{
		"update",
	}, rec.calls)
	require.Equal(t, rec.wrote.removePayload.Clients, []string{"testhost"})
}

func TestRemove_Data(t *testing.T) {
	setupStubs(t)

	rec := &removeRecorder{}

	_, err := service.remove(&Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		Content:       map[string]interface{}{"foo": "foo-value-1"},
	}, rec.ops())
	require.NoError(t, err)
	require.Equal(t, []string{
		"getItem",
		"update",
	}, rec.calls)
	require.Equal(t, rec.wrote.removePayload.Content, map[string]interface{}{"bar": "bar-value-1"})
}

func TestRemove_pruneData(t *testing.T) {
	type data chef.DataBagItem
	var existing data
	var remove data
	var want data
	rawExist := `{
				"foo": "foo-value-1",
				"bar": "bar-value-1",
				"baz": {
					"fuz": "fuz-value-1",
					"buz": "buz-value-1"
				}
			}`
	rawRemove := `{ "baz": { "fuz": "fuz-value-1" } }`
	rawWant := `{
				"foo": "foo-value-1",
				"bar": "bar-value-1",
				"baz": {
					"buz": "buz-value-1"
				}
			}`

	if err := json.Unmarshal([]byte(rawExist), &existing); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(rawRemove), &remove); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(rawWant), &want); err != nil {
		t.Fatal(err)
	}

	got, ok := pruneData(existing, remove)
	if !ok {
		t.Errorf("pruneData did not work")
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("pruneData did not work")
	}

}

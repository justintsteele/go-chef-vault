package vault

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/justintsteele/go-chef-vault/item_keys"
)

func TestUpdate_SwitchKeysMode(t *testing.T) {
	setupStubs(t)

	km := item_keys.KeysModeSparse
	pl := &Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		KeysMode:      &km,
	}

	got, err := service.Update(pl)
	if err != nil {
		t.Fatal(err)
	}

	want := &UpdateResponse{
		Response: Response{
			URI: fmt.Sprintf("%s/data/vault1", server.URL),
		},
		Data: &UpdateDataResponse{
			URI: fmt.Sprintf("%s/data/vault1/secret1", server.URL),
		},
		KeysURIs: []string{
			fmt.Sprintf("%s/data/vault1/secret1_keys", server.URL),
			fmt.Sprintf("%s/data/vault1/secret1_key_testhost", server.URL),
			fmt.Sprintf("%s/data/vault1/secret1_key_testhost3", server.URL),
			fmt.Sprintf("%s/data/vault1/secret1_key_pivotal", server.URL),
			fmt.Sprintf("%s/data/vault1/secret1_key_tester", server.URL),
			fmt.Sprintf("%s/data/vault1/secret1_key_testhost4", server.URL),
		},
	}

	if !reflect.DeepEqual(got.URI, want.URI) {
		t.Errorf("Update data URI = %v, want %v", got.URI, want.URI)
	}

	if !item_keys.EqualLists(got.KeysURIs, want.KeysURIs) {
		t.Errorf("Update data URI = %v, want %v", got.KeysURIs, want.KeysURIs)
	}
}

func TestUpdate_ResolveUpdateContent(t *testing.T) {
	setupStubs(t)

	payload, err := stubPayload([]string{"tester"}, []string{"testhost"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]interface{}
	updateContent := `{
						"foo": "fake-foo-value",
						"bar": "fake-bar-value-2"
					}`
	err = json.Unmarshal([]byte(updateContent), &raw)
	if err != nil {
		t.Fatal(err)
	}

	payload.Content = raw

	content, err := service.resolveUpdateContent(payload)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(content, payload.Content) {
		t.Errorf("payload.Content = %v, want %v", content, payload.Content)
	}

}

func TestUpdate_MergeKeyActors(t *testing.T) {
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

package vault

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

func TestBuildKeys_NoAdminsFails(t *testing.T) {
	keysMode := item_keys.KeysModeDefault
	secret, _ := item_keys.GenSecret(32)
	payload := &Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		Content:       nil,
		KeysMode:      &keysMode,
		SearchQuery:   nil,
		Admins:        []string{},
		Clients:       []string{},
	}
	_, err := service.buildKeys(payload, secret)
	if err == nil {
		t.Fatal("expected error when no admins resolve")
	}
}

func TestResolveSearchQuery_PreservesExisting(t *testing.T) {
	setupStubs(t)

	// send a payload with a nil query at a stubbed vault with a query to ensure the query is preserved
	payload, _ := stubPayload([]string{"tester"}, []string{"testhost", "testhost2", "testhost3", "testhost4"}, nil)

	keyState, err := service.loadKeysCurrentState(payload)
	if err != nil {
		t.Fatal(err)
	}
	finalQuery := item_keys.ResolveSearchQuery(keyState.SearchQuery, payload.SearchQuery)
	shouldQuery := "name:testhost*"
	if *finalQuery != shouldQuery {
		t.Fatalf("unexpected search query")
	}
}

func TestResolveSearchQuery_OverwriteWithNew(t *testing.T) {
	setupStubs(t)

	payload, err := stubPayload([]string{"tester"}, []string{"testhost"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	keyState, err := service.loadKeysCurrentState(payload)
	if err != nil {
		t.Fatal(err)
	}
	query := "name:testhost* AND chef_environment:development"
	payload.SearchQuery = &query

	resolvedQuery := item_keys.ResolveSearchQuery(keyState.SearchQuery, payload.SearchQuery)

	if !reflect.DeepEqual(resolvedQuery, payload.SearchQuery) {
		t.Errorf("payload.SearchQuery = %v, want %v", resolvedQuery, payload.SearchQuery)
	}
}

func TestMapKeys_SortsAndDedupes(t *testing.T) {
	input := map[string]chef.AccessKey{
		"testhost3": {},
		"tester":    {},
		"testhost":  {},
	}

	got := item_keys.MapKeys(input)

	want := []string{"tester", "testhost", "testhost3"}

	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Fatalf("missing key %q in %v", w, got)
		}
	}
}

func stubPayload(admins []string, clients []string, searchQuery *string) (*Payload, error) {
	content := `{"baz": "baz-value-1", "fuz": "fuz-value-2"}`
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		panic(err)
	}
	admins = append(admins, client.Auth.ClientName)
	clients = append(clients, "testhost")
	keysMode := item_keys.KeysModeDefault
	return &Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		Content:       raw,
		KeysMode:      &keysMode,
		SearchQuery:   searchQuery,
		Admins:        admins,
		Clients:       clients,
	}, nil
}

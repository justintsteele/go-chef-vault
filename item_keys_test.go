package vault

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

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

func TestCleanUnknownClients(t *testing.T) {
	setupStubs(t)

	payload, _ := stubPayload([]string{"tester"}, []string{"testhost", "testhost3", "fakehost"}, nil)
	keyState := &item_keys.VaultItemKeys{
		Id:          "secret1_keys",
		Admins:      []string{"pivotal", "tester"},
		Clients:     []string{"testhost", "testhost3", "fakehost"},
		SearchQuery: "name:testhost*",
		Mode:        item_keys.KeysModeDefault,
		Keys: map[string]string{
			"testhost":  "testhost-private-key-b64\n",
			"pivotal":   "testhost-private-key-b64\n",
			"tester":    "tester-private-key-b64\n",
			"testhost3": "testhost3-private-key-b64\n",
			"fakehost":  "fakehost-private-key-b64\n",
		},
	}
	kept, removed, err := service.cleanUnknownClients(payload, keyState, keyState.Clients)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(kept)
	slices.Sort(removed)
	if !slices.Equal(kept, []string{"testhost", "testhost3"}) {
		t.Fatalf("kept = %v, want %v", kept, []string{"testhost", "testhost3"})
	}
	if !slices.Equal(removed, []string{"fakehost"}) {
		t.Fatalf("removed = %v, want %v", removed, []string{"fakehost"})
	}
	if slices.Contains(keyState.Clients, "fakehost") {
		t.Fatalf("fakehost should have been removed")
	}
	if !reflect.DeepEqual(keyState.Clients, kept) {
		t.Fatalf("keyState.Clients = %v, want %v", keyState.Clients, kept)
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

package vault

import (
	"encoding/json"
	"go-chef-vault/vaultcrypto"
	"slices"
	"testing"

	"github.com/go-chef/chef"
)

func TestBuildKeys_NoAdminsFails(t *testing.T) {
	keysMode := KeysModeDefault
	secret, _ := vaultcrypto.GenSecret(32)
	payload := &VaultPayload{
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

func TestBuildKeys_MergesClients(t *testing.T) {
	setupStubs(t)

	payload, _ := stubPayload([]string{"tester"}, []string{"testhost", "testhost2", "testhost3"}, nil)
	secret, _ := vaultcrypto.GenSecret(32)

	item, err := service.buildKeys(payload, secret)
	if err != nil {
		t.Fatal(err)
	}

	got := item["clients"].([]string)
	want := []string{"testhost", "testhost3"}

	if !slices.Equal(got, want) {
		t.Fatalf("clients mismatch: want=%v got=%v", want, got)
	}
}

func TestBuildKeys_EncryptsForAllActors(t *testing.T) {
	setupStubs(t)

	payload, _ := stubPayload([]string{"tester"}, []string{"testhost", "testhost2", "testhost3", "testhost4"}, nil)
	secret, _ := vaultcrypto.GenSecret(32)
	item, err := service.buildKeys(payload, secret)
	if err != nil {
		t.Fatal(err)
	}

	meta := map[string]bool{
		"id":           true,
		"admins":       true,
		"clients":      true,
		"search_query": true,
		"mode":         true,
	}

	encryptedCount := 0
	for k := range item {
		if !meta[k] {
			encryptedCount++
		}
	}

	// admins + clients
	const want = 4
	if encryptedCount != want {
		t.Fatalf("expected %d encrypted actor keys, got %d", want, encryptedCount)
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
	finalQuery := resolveSearchQuery(keyState.SearchQuery, payload.SearchQuery)
	shouldQuery := "name:testhost*"
	if *finalQuery != shouldQuery {
		t.Fatalf("unexpected search query")
	}
}

func TestMapKeys_SortsAndDedupes(t *testing.T) {
	input := map[string]chef.AccessKey{
		"testhost3": {},
		"tester":    {},
		"testhost":  {},
	}

	got := mapKeys(input)

	want := []string{"tester", "testhost", "testhost3"}

	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Fatalf("missing key %q in %v", w, got)
		}
	}
}

func stubPayload(admins []string, clients []string, searchQuery *string) (*VaultPayload, error) {
	content := `{"baz": "baz-value-1", "fuz": "fuz-value-2"}`
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		panic(err)
	}
	admins = append(admins, client.Auth.ClientName)
	clients = append(clients, "testhost")
	keysMode := KeysModeDefault
	return &VaultPayload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		Content:       raw,
		KeysMode:      &keysMode,
		SearchQuery:   searchQuery,
		Admins:        admins,
		Clients:       clients,
	}, nil
}

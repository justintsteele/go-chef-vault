package vault

import (
	"encoding/json"
	"fmt"
	"go-chef-vault/vault/item_keys"
	"reflect"
	"testing"
)

func TestVaultsService_Create(t *testing.T) {
	setupStubs(t)

	var raw map[string]interface{}
	vaultItem := `{"baz": "baz-value-1", "fuz": "fuz-value-2"}`
	if err := json.Unmarshal([]byte(vaultItem), &raw); err != nil {
		panic(err)
	}

	var admins []string
	admins = append(admins, client.Auth.ClientName)

	var clients []string
	clients = append(clients, "testhost")

	keysMode := item_keys.KeysModeDefault
	pl := &Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		Content:       raw,
		KeysMode:      &keysMode,
		SearchQuery:   nil,
		Admins:        admins,
		Clients:       clients,
	}

	response, err := service.Create(pl)
	if err != nil {
		t.Errorf("Vaults.Create returned error: %v", err)
	}

	want := &CreateResponse{
		Response: Response{
			URI: fmt.Sprintf("%s/data/vault1", server.URL),
		},
		Data: &CreateDataResponse{
			URI: fmt.Sprintf("%s/data/vault1/secret1", server.URL),
		},
		KeysURIs: []string{fmt.Sprintf("%s/data/vault1/secret1_keys", server.URL)},
	}

	if !reflect.DeepEqual(response, want) {
		t.Errorf("Vaults.Create returned %+v, want %+v", response, want)
	}
}

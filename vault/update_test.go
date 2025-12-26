package vault

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestVaultsService_Update(t *testing.T) {
	setup()
	defer teardown()

	cleanupDecrypt := stubVaultItemKeyDecrypt(t)
	defer cleanupDecrypt()

	cleanupEncrypt := stubVaultItemKeyEncrypt(t)
	defer cleanupEncrypt()

	stubMuxGetItem(t)
	stubMuxCreate(t)

	var raw map[string]interface{}
	vaultItem := `{"baz": "baz-value-1", "fuz": "fuz-value-2"}`
	if err := json.Unmarshal([]byte(vaultItem), &raw); err != nil {
		panic(err)
	}

	var admins []string
	admins = append(admins, client.Auth.ClientName)

	var clients []string
	clients = append(clients, "testhost")

	keysMode := KeysModeDefault
	pl := &VaultPayload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		Content:       raw,
		KeysMode:      &keysMode,
		SearchQuery:   nil,
		Admins:        admins,
		Clients:       clients,
	}

	response, err := service.Update(pl)
	if err != nil {
		t.Errorf("Vaults.Create returned error: %v", err)
	}

	want := &UpdateResponse{
		VaultResponse: VaultResponse{
			URI: fmt.Sprintf("%s/data/vault1", server.URL),
		},
		Data: &UpdateDataResponse{
			URI: fmt.Sprintf("%s/data/vault1/secret1", server.URL),
		},
		KeysURIs: []string{fmt.Sprintf("%s/data/vault1/secret1_keys", server.URL)},
	}

	if !reflect.DeepEqual(response, want) {
		t.Errorf("Vaults.Update returned %+v, want %+v", response, want)
	}
}

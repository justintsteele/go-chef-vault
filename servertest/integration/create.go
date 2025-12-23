package integration

import (
	"encoding/json"
	"fmt"

	"github.com/justintsteele/go-chef-vault/vault"
)

func createVault(service *vault.Service) (res *vault.CreateResponse, err error) {
	var raw map[string]interface{}
	vaultItem := `{"baz": "baz-value-1", "fuz": "fuz-value-2"}`

	if err := json.Unmarshal([]byte(vaultItem), &raw); err != nil {
		panic(err)
	}

	var admins []string
	admins = append(admins, service.Client.Auth.ClientName)

	// purposefully omit KeysMode so we can test behavior of changing the mode later
	// purposefully omit SearchQuery because in goiardi we don't have other clients to search
	pl := &vault.VaultPayload{
		VaultName:     vaultName,
		VaultItemName: vaultItemName,
		Content:       raw,
		KeysMode:      nil,
		SearchQuery:   nil,
		Admins:        admins,
		Clients:       []string{},
	}

	res, err = service.Create(pl)
	if err != nil {
		fmt.Printf("Vault %s/%s already exists\n", vaultName, vaultItemName)
	}

	return
}

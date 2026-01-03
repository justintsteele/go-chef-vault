package integration

import (
	"encoding/json"
	"go-chef-vault/vault/item_keys"

	"github.com/justintsteele/go-chef-vault/vault"
)

func (i *IntegrationService) updateContent() (result *vault.UpdateResponse, err error) {
	var raw map[string]interface{}
	vaultItem := `{"baz": "baz-value-1", "fuz": "fuz-value-2", "foo": "foo-value-3"}`

	if err = json.Unmarshal([]byte(vaultItem), &raw); err != nil {
		return
	}

	// Here we add a search query to the vault so it gets picked up by the refresh step.
	query := "name:testhost*"
	var admins []string
	admins = append(admins, i.Service.Client.Auth.ClientName)

	keysMode := item_keys.KeysModeSparse
	pl := &vault.Payload{
		VaultName:     vaultName,
		VaultItemName: vaultItemName,
		Content:       raw,
		KeysMode:      &keysMode,
		SearchQuery:   &query,
		Admins:        admins,
		Clients:       []string{},
	}

	result, err = i.Service.Update(pl)
	if err != nil {
		return
	}

	return
}

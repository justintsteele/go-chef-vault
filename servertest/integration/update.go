package integration

import (
	"encoding/json"

	"github.com/justintsteele/go-chef-vault/vault"
)

func updateContent(service *vault.Service) (result *vault.UpdateResponse, err error) {
	var raw map[string]interface{}
	vaultItem := `{"baz": "baz-value-1", "fuz": "fuz-value-2", "foo": "foo-value-3"}`

	if err = json.Unmarshal([]byte(vaultItem), &raw); err != nil {
		return
	}

	var admins []string
	admins = append(admins, service.Client.Auth.ClientName)

	pl := &vault.VaultPayload{
		VaultName:     vaultName,
		VaultItemName: vaultItemName,
		Content:       raw,
		KeysMode:      nil,
		SearchQuery:   nil,
		Admins:        admins,
		Clients:       []string{},
	}

	result, err = service.Update(pl)
	if err != nil {
		return
	}

	return
}

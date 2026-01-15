package integration

import (
	"encoding/json"

	"github.com/justintsteele/go-chef-vault"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

func (i *IntegrationService) updateContent() (result *vault.UpdateResponse, err error) {
	var raw map[string]interface{}
	vaultItem := `{
"baz": "baz-value-1", 
"foo": "foo-value-3",
"fuz": {
	  "faz": "faz-value-4",
	  "buz": "buz-value-5",
	  "boz": "boz-value-6"
	}
}`
	// Here we add a new node and client so that the search query we added in update has something new to find.
	Must(i.createClients(newNodeName))
	Must(i.createClients(newNodeName + "1"))
	Must(i.createClients(fakeNodeName))

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
		Clients:       []string{fakeNodeName},
	}

	result, err = i.Service.Update(pl)
	if err != nil {
		return
	}

	return
}

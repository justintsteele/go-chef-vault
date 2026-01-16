package integration

import (
	"encoding/json"

	"github.com/justintsteele/go-chef-vault"
)

func (i *IntegrationService) createVault() (res *vault.CreateResponse, err error) {
	var raw map[string]interface{}
	vaultItem := `{"baz": "baz-value-1", "fuz": "fuz-value-2"}`

	if err := json.Unmarshal([]byte(vaultItem), &raw); err != nil {
		panic(err)
	}

	var admins []string
	admins = append(admins, i.Service.Client.Auth.ClientName)

	// purposefully omit KeysMode so we can test behavior of changing the mode later
	pl := &vault.Payload{
		VaultName:     vaultName,
		VaultItemName: vaultItemName,
		Content:       raw,
		KeysMode:      nil,
		SearchQuery:   nil,
		Admins:        admins,
		Clients:       []string{},
	}

	res, _ = i.Service.Create(pl)

	// create a second vault so we can see rotate all keys do more than one vault later
	pl2 := &vault.Payload{
		VaultName:     "go-vault2",
		VaultItemName: "secret2",
		Content:       raw,
		Admins:        admins,
		Clients:       []string{},
	}

	_, _ = i.Service.Create(pl2)

	return
}

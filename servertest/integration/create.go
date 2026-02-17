package integration

import (
	"encoding/json"
	"fmt"

	vault "github.com/justintsteele/go-chef-vault"
)

func createVault() Scenario {
	return Scenario{
		Name: "Create",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}

			var raw map[string]interface{}
			vaultItem := `{"baz": "baz-value-1", "fuz": "fuz-value-2"}`
			_ = json.Unmarshal([]byte(vaultItem), &raw)

			admins := []string{i.Service.Client.Auth.ClientName}

			pl := &vault.Payload{
				VaultName:     vaultName,
				VaultItemName: vaultItemName,
				Content:       raw,
				Admins:        admins,
			}

			_, err := i.Service.Create(pl)
			sr.assertNoError(fmt.Sprintf("create vault %s", vaultName), err)

			pl2 := &vault.Payload{
				VaultName:     vault2Name,
				VaultItemName: vault2ItemName,
				Content:       raw,
				Admins:        admins,
			}

			_, err = i.Service.Create(pl2)
			sr.assertNoError(fmt.Sprintf("create vault %s", vault2Name), err)

			_, err = i.Service.Create(nil)
			sr.assertError(fmt.Sprintf("nil payload: %v", err), err)

			_, err = i.Service.Create(&vault.Payload{VaultItemName: "secret4"})
			sr.assertError(fmt.Sprintf("missing vault name: %v", err), err)

			return sr
		},
	}
}

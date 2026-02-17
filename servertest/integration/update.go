package integration

import (
	"encoding/json"
	"fmt"

	"github.com/justintsteele/go-chef-vault"
	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

type KeyMode int

const (
	SparseOnly KeyMode = iota
	DefaultOnly
)

func updateSparse() Scenario {
	return Scenario{
		Name: "Update (Sparse Keys)",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}

			// fixtures: if these fail, scenario can't proceed meaningfully
			// Here we add a new node and client so that the search query we added in update has something new to find.
			if err := i.createClients(newNodeName); err != nil {
				sr.assertNoError(fmt.Sprintf("Create %s", newNodeName), err)
			}

			if err := i.createClients(newNodeName + "1"); err != nil {
				sr.assertNoError(fmt.Sprintf("Create %s", newNodeName+"1"), err)
			}

			if err := i.createClients(fakeNodeName); err != nil {
				sr.assertNoError(fmt.Sprintf("Create %s", fakeNodeName), err)
			}

			// setting up a new content payload
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
			if err := json.Unmarshal([]byte(vaultItem), &raw); err != nil {
				sr.assertNoError("unmarshal vaultItem", err)
				return sr
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

			_, err := i.Service.Update(pl)
			sr.assertNoError("update vaultItem", err)

			db, err := i.Service.GetItem(vaultName, vaultItemName)
			sr.assertNoError("get vaultItem", err)

			dbi, _ := item.DataBagItemMap(db)
			raw["id"] = vaultItemName
			sr.assertEqual("updated content", raw, dbi)

			keys, _ := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_keys")
			keysDbi, _ := item.DataBagItemMap(keys)

			assertSparseKeysOnly(sr, i.Service, vaultName, vaultItemName, keysDbi, "clients")
			assertSparseKeysOnly(sr, i.Service, vaultName, vaultItemName, keysDbi, "admins")

			_, err = i.Service.Update(nil)
			sr.assertError(fmt.Sprintf("nil payload: %v", err), err)

			return sr
		},
	}
}

func updateDefault() Scenario {
	return Scenario{
		Name: "Update (Default Keys)",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}
			keysMode := item_keys.KeysModeDefault
			pl := &vault.Payload{
				VaultName:     vaultName,
				VaultItemName: vaultItemName,
				KeysMode:      &keysMode,
			}

			_, err := i.Service.Update(pl)
			sr.assertNoError("update keys mode to default", err)

			keys, _ := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_keys")
			keysDbi, _ := item.DataBagItemMap(keys)

			assertDefaultKeysOnly(sr, i.Service, vaultName, vaultItemName, keysDbi, "clients")
			assertDefaultKeysOnly(sr, i.Service, vaultName, vaultItemName, keysDbi, "admins")

			_, err = i.Service.Update(&vault.Payload{VaultItemName: "secret4"})
			sr.assertError(fmt.Sprintf("missing vault name: %v", err), err)

			return sr
		},
	}
}

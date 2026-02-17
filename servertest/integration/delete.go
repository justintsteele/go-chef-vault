package integration

import (
	"fmt"

	"github.com/justintsteele/go-chef-vault/cheferr"
)

func deleteItem() Scenario {
	return Scenario{
		Name: "Delete Item",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}

			vaultItem, err := i.Service.Client.DataBags.ListItems(vaultName)
			sr.assert(fmt.Sprintf("pre-delete %s exists", vaultItemName), (*vaultItem)[vaultItemName] != "", err)

			_, err = i.Service.DeleteItem(vaultName, vaultItemName)
			sr.assertNoError(fmt.Sprintf("delete item %s/%s", vaultName, vaultItemName), err)

			postVaultItem, err := i.Service.Client.DataBags.ListItems(vaultName)
			sr.assert(fmt.Sprintf("post-delete %s does not exist", vaultItemName), (*postVaultItem)[vaultItemName] == "", err)

			_, err = i.Service.DeleteItem("", vaultItemName)
			sr.assertError(fmt.Sprintf("missing vault name: %v", err), err)

			return sr
		},
	}
}

func deleteVault(name, item string) Scenario {
	return Scenario{
		Name: "Delete Vault",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}

			_, err := i.Service.Delete(name)
			sr.assertNoError(fmt.Sprintf("delete %s", name), err)

			_, err = i.Service.Client.DataBags.GetItem(name, item)
			sr.assert(fmt.Sprintf("%s does not exist", name), cheferr.IsNotFound(err), fmt.Errorf("%s still exists", name))

			_, err = i.Service.Delete("")
			sr.assertError(fmt.Sprintf("empty vault name: %v", err), err)

			return sr
		},
	}
}

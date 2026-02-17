package integration

import (
	"encoding/json"
	"fmt"

	"github.com/justintsteele/go-chef-vault/item"
)

func getVault() Scenario {
	return Scenario{
		Name: "Get Item",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}

			res, err := i.Service.GetItem(vaultName, vaultItemName)
			sr.assertNoError(fmt.Sprintf("Get Vault: %s", vaultName), err)

			dbi, dbierr := item.DataBagItemMap(res)
			sr.assertNoError("Data Bag Item Mapped", dbierr)

			var raw map[string]interface{}
			rawItem := `{"id": "secret1", "baz": "baz-value-1", "fuz": "fuz-value-2"}`
			if err := json.Unmarshal([]byte(rawItem), &raw); err != nil {
				sr.assertNoError("Error unmarshalling raw item", err)
				return sr
			}

			if dbierr == nil {
				sr.assertEqual("retrieved content", raw, dbi)
			}

			_, err = i.Service.GetItem(vaultName, "")
			sr.assertError(fmt.Sprintf("empty vault item name: %v", err), err)

			return sr
		},
	}
}

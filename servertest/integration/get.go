package integration

import (
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

			if dbierr == nil {
				val, ok := dbi["baz"]
				if !ok {
					sr.assert("Decrypted value does not contain key 'baz'", false, fmt.Errorf("key 'baz' not found in decrypted item"))
				} else if s, ok := val.(string); !ok {
					sr.assert("Decrypted value type is not a string", false, fmt.Errorf("key 'baz' has type %T, expected string", val))
				} else {
					sr.assertEqual("Decrypted value matches expected", s, "baz-value-1")
				}
			}

			return sr
		},
	}
}

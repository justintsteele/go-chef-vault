package integration

import (
	"errors"
	"fmt"

	"github.com/justintsteele/go-chef-vault"
	"github.com/justintsteele/go-chef-vault/item"
)

func rotate() Scenario {
	return Scenario{
		Name: "Rotate Keys",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}

			pl := &vault.Payload{
				VaultName:     vaultName,
				VaultItemName: vaultItemName,
				CleanUnknown:  true,
			}

			preUserKey, _ := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_key_"+goiardiUser)
			preUserDbi, _ := item.DataBagItemMap(preUserKey)

			_, err := i.Service.RotateKeys(pl)
			sr.assertNoError("rotate keys", err)

			postUserKey, _ := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_key_"+goiardiUser)
			postUserDbi, _ := item.DataBagItemMap(postUserKey)

			ok := preUserDbi[goiardiUser] != postUserDbi[goiardiUser]
			sr.assert(fmt.Sprintf("%s key rotated", goiardiUser), ok, errors.New("user key not rotated"))

			_, err = i.Service.GetItem(vaultName, vaultItemName)
			sr.assertNoError("get vault item", err)

			_, err = i.Service.RotateKeys(nil)
			sr.assertError(fmt.Sprintf("nil payload: %v", err), err)

			return sr
		},
	}
}

func rotateAll() Scenario {
	return Scenario{
		Name: "Rotate All Keys",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}

			result, err := i.Service.RotateAllKeys()
			sr.assertNoError("rotate all keys", err)
			sr.assertEqual("number of keys rotated", len(result), 2)

			return sr
		},
	}
}

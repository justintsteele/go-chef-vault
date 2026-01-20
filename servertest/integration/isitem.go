package integration

import (
	"fmt"

	"github.com/go-chef/chef"
	vault "github.com/justintsteele/go-chef-vault"
	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

const (
	dataBagName     string = "databag1"
	dataBagItemName string = "item1"
	encrDataBagName string = "encrypt1"
	encrDataBagItem string = "encrypted_item1"
)

func isItem() Scenario {
	return Scenario{
		Name: "IsVault?",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}

			// fixtures: if these fail, scenario can't proceed meaningfully
			if err := i.createDataBag(); err != nil {
				sr.assertNoError("Create Data Bag", err)
				return sr
			}
			sr.assertNoError("Create Data Bag", nil)

			if err := i.createEncryptedDataBag(); err != nil {
				sr.assertNoError("Create Encrypted Data Bag", err)
				return sr
			}
			sr.assertNoError("Create Encrypted Data Bag", nil)

			assertVaultState(sr, i.Service, vaultName, vaultItemName, true, vault.DataBagItemTypeVault)
			assertVaultState(sr, i.Service, encrDataBagName, encrDataBagItem, false, vault.DataBagItemTypeEncrypted)
			assertVaultState(sr, i.Service, dataBagName, dataBagItemName, false, vault.DataBagItemTypeNormal)

			return sr
		},
	}
}

func assertVaultState(sr *ScenarioResult, svc *vault.Service, bag, item string, wantVault bool, wantType vault.DataBagItemType) {
	isVault, _ := svc.IsVault(bag, item)
	var vaultStr string
	if !wantVault {
		vaultStr = "not "
	}

	sr.assertEqual(fmt.Sprintf("%s/%s is %sa vault", bag, item, vaultStr), wantVault, isVault)

	itemType, _ := svc.ItemType(bag, item)
	sr.assertEqual(fmt.Sprintf("%s/%s is item type [%s]", bag, item, itemType), wantType, itemType)
}

func (i *IntegrationService) createDataBag() error {
	// create a data bag
	dataBag := chef.DataBag{
		Name: dataBagName,
	}
	_, err := i.Service.Client.DataBags.Create(&dataBag)
	if err != nil {
		return err
	}

	dataBagItem := map[string]any{
		"id":  dataBagItemName,
		"foo": "foo-value-plaintext",
	}

	if err := i.Service.Client.DataBags.CreateItem(dataBag.Name, &dataBagItem); err != nil {
		return err
	}
	return nil
}

func (i *IntegrationService) createEncryptedDataBag() error {
	// create an encrypted data bag
	eDataBag := chef.DataBag{
		Name: encrDataBagName,
	}
	_, err := i.Service.Client.DataBags.Create(&eDataBag)
	if err != nil {
		return err
	}
	secret, err := item_keys.GenSecret(32)
	if err != nil {
		return err
	}

	eDataBagItem := map[string]any{
		"bar": "bar-value-encrypted",
	}

	eDB, err := item.Encrypt(encrDataBagItem, eDataBagItem, secret)
	if err != nil {
		return err
	}

	if err := i.Service.Client.DataBags.CreateItem(eDataBag.Name, &eDB); err != nil {
		return err
	}
	return nil
}

func (i *IntegrationService) deleteDataBags() error {
	_, err := i.Service.Client.DataBags.Delete(dataBagName)
	if err != nil {
		return err
	}
	_, err = i.Service.Client.DataBags.Delete(encrDataBagName)
	if err != nil {
		return err
	}
	return nil
}

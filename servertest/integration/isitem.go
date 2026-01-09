package integration

import (
	"fmt"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

const (
	dataBagName     string = "databag1"
	dataBagItemName string = "item1"
	encrDataBagName string = "encrypt1"
	encrDataBagItem string = "encrypted_item1"
)

func (i *IntegrationService) isVault() (string, error) {
	// these are just fixtures so we can get real results
	if err := i.createDataBag(); err != nil {
		return "", err
	}

	if err := i.createEncryptedDataBag(); err != nil {
		return "", err
	}

	bags := map[string]string{
		vaultName:       vaultItemName,
		encrDataBagName: encrDataBagItem,
		dataBagName:     dataBagItemName,
	}

	for k, v := range bags {
		vResult, err := i.Service.IsVault(k, v)
		if err != nil {
			return "", err
		}

		if vResult {
			fmt.Printf("%s/%s is a vault\n", k, v)
		} else {
			fmt.Printf("%s/%s is not a vault\n", k, v)
			itemTypeResult, err := i.Service.ItemType(k, v)
			if err != nil {
				return "", err
			}
			fmt.Printf("%s/%s is a %s data bag\n", k, v, itemTypeResult)
		}

	}
	return "", nil
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
	_, derr := i.Service.Client.DataBags.Delete(dataBagName)
	if derr != nil {
		return derr
	}
	fmt.Printf("deleted data bag %s\n", dataBagName)

	_, eerr := i.Service.Client.DataBags.Delete(encrDataBagName)
	if eerr != nil {
		return eerr
	}
	fmt.Printf("deleted encrypted data bag %s\n", encrDataBagName)

	return nil
}

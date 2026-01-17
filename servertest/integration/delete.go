package integration

import (
	"github.com/justintsteele/go-chef-vault"
)

func (i *IntegrationService) deleteVault() (result *vault.DeleteResponse, err error) {
	result, err = i.Service.Delete(vaultName)
	if err != nil {
		return
	}

	result, err = i.Service.Delete("go-vault2")
	if err != nil {
		return
	}
	return
}

func (i *IntegrationService) deleteItem() (result *vault.DeleteResponse, err error) {
	result, err = i.Service.DeleteItem(vaultName, vaultItemName)
	if err != nil {
		return
	}

	return
}

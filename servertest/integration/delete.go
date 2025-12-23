package integration

import (
	"github.com/justintsteele/go-chef-vault/vault"
)

func deleteVault(service *vault.Service) (result *vault.DeleteResponse, err error) {
	result, err = service.Delete(vaultName)
	if err != nil {
		return
	}

	return
}

func deleteItem(service *vault.Service) (result *vault.DeleteResponse, err error) {
	result, err = service.DeleteItem(vaultName, vaultItemName)
	if err != nil {
		return
	}

	return
}

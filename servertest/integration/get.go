package integration

import (
	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/vault"
)

func getVault(service *vault.Service) (result chef.DataBagItem, err error) {
	result, err = service.GetItem(vaultName, vaultItemName)
	if err != nil {
		return
	}

	return
}

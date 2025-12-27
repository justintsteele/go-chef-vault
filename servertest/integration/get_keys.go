package integration

import (
	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/vault"
)

func getKeys(service *vault.Service) (result chef.DataBagItem, err error) {
	result, err = service.Client.DataBags.GetItem(vaultName, vaultItemName+"_keys")
	if err != nil {
		return
	}

	return
}

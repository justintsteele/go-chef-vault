package integration

import (
	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/vault"
)

func list(service *vault.Service) (result *chef.DataBagListResult, err error) {
	result, err = service.List()
	if err != nil {
		return
	}

	return
}

func listItems(service *vault.Service) (result *chef.DataBagListResult, err error) {
	result, err = service.ListItems(vaultName)
	if err != nil {
		return
	}

	return
}

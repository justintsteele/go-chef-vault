package integration

import (
	"github.com/go-chef/chef"
)

func (i *IntegrationService) getVault() (result chef.DataBagItem, err error) {
	result, err = i.Service.GetItem(vaultName, vaultItemName)
	if err != nil {
		return
	}

	return
}

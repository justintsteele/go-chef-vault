package integration

import (
	"github.com/go-chef/chef"
)

func (i *IntegrationService) getKeys() (result chef.DataBagItem, err error) {
	result, err = i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_keys")
	if err != nil {
		return
	}

	return
}

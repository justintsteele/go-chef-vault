package integration

import (
	"github.com/go-chef/chef"
)

func (i *IntegrationService) list() (result *chef.DataBagListResult, err error) {
	result, err = i.Service.List()
	if err != nil {
		return
	}

	return
}

func (i *IntegrationService) listItems() (result *chef.DataBagListResult, err error) {
	result, err = i.Service.ListItems(vaultName)
	if err != nil {
		return
	}

	return
}

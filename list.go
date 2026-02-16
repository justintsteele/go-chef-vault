package vault

import (
	"strings"

	"github.com/go-chef/chef"
)

// List returns a list of vaults on the server.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#get-24
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_list.rb
func (s *Service) List() (*chef.DataBagListResult, error) {
	dbl, err := s.Client.DataBags.List()
	if err != nil {
		return nil, err
	}

	list := chef.DataBagListResult{}

	for bag, url := range *dbl {
		isVault, err := s.bagIsVault(bag)
		if err != nil {
			return nil, err
		}

		if isVault {
			list[bag] = url

		}
	}

	return &list, nil
}

// ListItems returns a list of the items in a vault.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#get-25
func (s *Service) ListItems(vaultName string) (*chef.DataBagListResult, error) {
	if vaultName == "" {
		return nil, ErrMissingVaultName
	}

	dbl, err := s.Client.DataBags.ListItems(vaultName)
	if err != nil {
		return nil, err
	}

	items := chef.DataBagListResult{}

	for item, url := range *dbl {
		if strings.HasSuffix(item, "_keys") {
			continue
		} else if strings.Contains(item, "_key_") {
			continue
		} else {
			items[item] = url
		}
	}

	return &items, nil
}

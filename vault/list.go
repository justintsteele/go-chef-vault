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
func (s *Service) List() (data *chef.DataBagListResult, err error) {
	dbl, err := s.Client.DataBags.List()
	if err != nil {
		return
	}

	list := chef.DataBagListResult{}

	for bag, url := range *dbl {
		if s.bagIsVault(bag) {
			list[bag] = url
		}
	}

	return &list, nil
}

// ListItems returns a list of the items in a vault.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#get-25
func (s *Service) ListItems(name string) (data *chef.DataBagListResult, err error) {
	dbl, err := s.Client.DataBags.ListItems(name)
	if err != nil {
		return
	}

	items := chef.DataBagListResult{}

	for item, url := range *dbl {
		if strings.HasSuffix(item, "_keys") {
			continue
		} else {
			items[item] = url
		}
	}

	return &items, nil
}

// bagIsVault returns bool of whether the specified data bag a vault.
//
// References:
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_base.rb#L51
func (s *Service) bagIsVault(bagName string) bool {
	rawItems, err := s.Client.DataBags.ListItems(bagName)
	if err != nil {
		return false
	}

	items := *rawItems

	for item := range items {
		if strings.HasSuffix(item, "_keys") {
			base := strings.TrimSuffix(item, "_keys")
			if _, ok := items[base]; ok {
				return true
			}
		}
	}
	return false
}

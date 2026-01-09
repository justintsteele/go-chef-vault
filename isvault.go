package vault

// IsVault determines whether the data bag item is a vault.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#get-24
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_isvault.rb
func (s *Service) IsVault(vaultName string, vaultItem string) (bool, error) {
	itemType, err := s.ItemType(vaultName, vaultItem)
	if err != nil {
		return false, err
	}

	return itemType == DataBagItemTypeVault, nil
}

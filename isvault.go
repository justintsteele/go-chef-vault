package vault

// IsVault determines whether the data bag item is a vault.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#get-24
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_isvault.rb
func (s *Service) IsVault(vaultName, vaultItem string) (bool, error) {
	pl := &Payload{
		VaultName:     vaultName,
		VaultItemName: vaultItem,
	}

	if err := pl.validatePayload(); err != nil {
		return false, err
	}

	itemType, err := s.ItemType(pl.VaultName, pl.VaultItemName)
	if err != nil {
		return false, err
	}

	return itemType == DataBagItemTypeVault, nil
}

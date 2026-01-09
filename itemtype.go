package vault

type DataBagItemType string

const (
	DataBagItemTypeVault     DataBagItemType = "vault"
	DataBagItemTypeEncrypted DataBagItemType = "encrypted"
	DataBagItemTypeNormal    DataBagItemType = "normal"
)

// ItemType determines whether the data bag item is a vault, encrypted data bag, or a normal data bag item.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#get-24
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_itemtype.rb
func (s *Service) ItemType(vaultName, vaultItem string) (DataBagItemType, error) {
	isVault, err := s.bagIsVault(vaultName)
	if err != nil {
		return "", err
	}

	if isVault {
		return DataBagItemTypeVault, nil
	}

	encrypted, err := s.bagItemIsEncrypted(vaultName, vaultItem)
	if err != nil {
		return "", err
	}

	if encrypted {
		return DataBagItemTypeEncrypted, nil
	}

	return DataBagItemTypeNormal, nil
}

package vault

type ItemType string

const (
	ItemTypeVault     ItemType = "vault"
	ItemTypeEncrypted ItemType = "encrypted"
	ItemTypeNormal    ItemType = "normal"
)

// ItemType determines whether the data bag item is a vault, encrypted data bag, or a normal data bag item.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#get-24
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_itemtype.rb
func (s *Service) ItemType(vaultName, vaultItem string) (ItemType, error) {
	encrypted, err := s.bagItemIsEncrypted(vaultName, vaultItem)
	if err != nil {
		return "", err
	}

	if !encrypted {
		return ItemTypeNormal, nil
	}

	isVault, err := s.bagIsVault(vaultName)
	if err != nil {
		return "", err
	}

	if isVault {
		return ItemTypeVault, nil
	}

	return ItemTypeEncrypted, nil
}

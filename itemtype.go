package vault

// DataBagItemType represents the classification of a Chef data bag item as determined by Chef-Vault semantics.
type DataBagItemType string

const (
	// DataBagItemTypeVault indicates the item is a Chef Vault.
	DataBagItemTypeVault DataBagItemType = "vault"

	// DataBagItemTypeEncrypted indicates the item is an encrypted data bag item containing
	// encrypted application data.
	DataBagItemTypeEncrypted DataBagItemType = "encrypted"

	// DataBagItemTypeNormal indicates the item is a standard Chef data bag item
	// with no Chef-Vault or encrypted data bag semantics applied.
	DataBagItemTypeNormal DataBagItemType = "normal"
)

// ItemType determines whether the data bag item is a vault, encrypted data bag, or a normal data bag item.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#get-24
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_itemtype.rb
func (s *Service) ItemType(vaultName, vaultItem string) (DataBagItemType, error) {
	pl := &Payload{
		VaultName:     vaultName,
		VaultItemName: vaultItem,
	}

	if err := pl.validatePayload(); err != nil {
		return "", err
	}

	isVault, err := s.bagIsVault(pl.VaultName)
	if err != nil {
		return "", err
	}

	if isVault {
		return DataBagItemTypeVault, nil
	}

	encrypted, err := s.bagItemIsEncrypted(pl.VaultName, pl.VaultItemName)
	if err != nil {
		return "", err
	}

	if encrypted {
		return DataBagItemTypeEncrypted, nil
	}

	return DataBagItemTypeNormal, nil
}

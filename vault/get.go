package vault

import (
	"fmt"
	"go-chef-vault/vault/item"
	"go-chef-vault/vault/item_keys"

	"github.com/go-chef/chef"
)

// GetItem returns the decrypted items in the vault.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#get-26
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_show.rb
func (s *Service) GetItem(vaultName, vaultItem string) (chef.DataBagItem, error) {
	// Fetch raw data bag items
	rawItem, err := s.Client.DataBags.GetItem(vaultName, vaultItem)
	if err != nil {
		return nil, err
	}

	rawKeys, err := s.Client.DataBags.GetItem(vaultName, vaultItem+"_keys")
	if err != nil {
		return nil, err
	}

	itemMap, err := item_keys.DataBagItemMap(rawItem)
	if err != nil {
		return nil, err
	}

	keysMap, err := item_keys.DataBagItemMap(rawKeys)
	if err != nil {
		return nil, err
	}

	var actor = s.Client.Auth.ClientName
	var actorKey string
	keyMode := keysMap["mode"]
	switch keyMode {
	case "default":
		publicKey, ok := keysMap[actor]
		if !ok {
			return nil, fmt.Errorf("%s/%s is not encrypted with your public key", vaultName, vaultItem)
		}
		actorKey = publicKey.(string)
	case "sparse":
		rawSparseKey, err := s.Client.DataBags.GetItem(vaultName, vaultItem+"_key_"+actor)
		if err != nil {
			return nil, fmt.Errorf("%s/%s is not encrypted with your public key", vaultName, vaultItem)
		}
		sparseKeyMap, err := item_keys.DataBagItemMap(rawSparseKey)
		if err != nil {
			return nil, err
		}
		actorKey = sparseKeyMap[actor].(string)
	}

	aesKey, err := item_keys.DeriveAESKeyForVault(
		actorKey,
		s.Client.Auth.PrivateKey,
	)
	if err != nil {
		return nil, err
	}

	evs := make(map[string]item.EncryptedValue)

	for dbi, val := range itemMap {
		if dbi == "id" {
			continue
		}

		m, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("item %s is not an encrypted value", dbi)
		}

		evs[dbi] = item.EncryptedValue{
			EncryptedData: m["encrypted_data"].(string),
			IV:            m["iv"].(string),
			AuthTag:       m["auth_tag"].(string),
			Version:       int(m["version"].(float64)),
			Cipher:        m["cipher"].(string),
		}
	}

	vault := &item.VaultItem{
		Id:    vaultItem,
		Items: evs,
	}

	return vault.Decrypt(aesKey)
}

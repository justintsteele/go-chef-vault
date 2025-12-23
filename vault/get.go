package vault

import (
	"fmt"

	"github.com/go-chef/chef"
)

// GetItem returns the decrypted items in the vault
//
//	Chef API Docs: https://docs.chef.io/api_chef_server/#get-26
//	Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_show.rb
func (v *Service) GetItem(vaultName, vaultItem string) (chef.DataBagItem, error) {
	// Fetch raw data bag items
	rawItem, err := v.Client.DataBags.GetItem(vaultName, vaultItem)
	if err != nil {
		return nil, err
	}

	rawKeys, err := v.Client.DataBags.GetItem(vaultName, vaultItem+"_keys")
	if err != nil {
		return nil, err
	}

	itemMap, err := dataBagItemMap(rawItem)
	if err != nil {
		return nil, err
	}

	keysMap, err := dataBagItemMap(rawKeys)
	if err != nil {
		return nil, err
	}

	var actor = v.Client.Auth.ClientName
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
		rawSparseKey, err := v.Client.DataBags.GetItem(vaultName, vaultItem+"_key_"+actor)
		if err != nil {
			return nil, fmt.Errorf("%s/%s is not encrypted with your public key", vaultName, vaultItem)
		}
		sparseKeyMap, err := dataBagItemMap(rawSparseKey)
		if err != nil {
			return nil, err
		}
		actorKey = sparseKeyMap[actor].(string)
	}

	if v.authorize == nil {
		v.authorize = v.authorizeVaultItem
	}

	aesKey, err := v.authorize(actorKey)
	if err != nil {
		return nil, err
	}

	evs := make(map[string]EncryptedValue)
	for item, val := range itemMap {
		if item == "id" {
			continue
		}

		m, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("item %s is not an encrypted value", item)
		}
		evs[item] = EncryptedValue{
			EncryptedData: m["encrypted_data"].(string),
			IV:            m["iv"].(string),
			AuthTag:       m["auth_tag"].(string),
			Version:       int(m["version"].(float64)),
			Cipher:        m["cipher"].(string),
		}
	}

	vault := &VaultItem{
		Id:    vaultItem,
		Items: evs,
	}

	return vault.decrypt(aesKey)
}

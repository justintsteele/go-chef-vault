package item

import (
	"encoding/json"

	chefcrypto "github.com/bhoriuchi/go-chef-crypto"
	"github.com/go-chef/chef"
)

// Encrypt creates the encrypted data bag item for a vault.
func Encrypt(vaultItemName string, content map[string]interface{}, secret []byte) (chef.DataBagItem, error) {
	item := make(map[string]any)
	item["id"] = vaultItemName

	for k, v := range content {
		plaintext, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		encrypted, err := chefcrypto.Encrypt(secret, plaintext, 3)
		if err != nil {
			return nil, err
		}
		item[k] = encrypted
	}

	return item, nil
}

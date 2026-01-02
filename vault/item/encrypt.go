package item

import (
	"encoding/json"

	chefcrypto "github.com/bhoriuchi/go-chef-crypto"
	"github.com/go-chef/chef"
)

var EncryptFunc = chefcrypto.Encrypt

// Encrypt creates the data half of the vault by creating an encrypted data bag item
func Encrypt(vaultItemName string, content map[string]interface{}, secret []byte) (chef.DataBagItem, error) {
	item := make(map[string]any)
	item["id"] = vaultItemName

	for k, v := range content {
		plaintext, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		encrypted, err := EncryptFunc(secret, plaintext, 3)
		if err != nil {
			return nil, err
		}
		item[k] = encrypted
	}

	return item, nil
}

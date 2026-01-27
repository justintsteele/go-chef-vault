package item

import (
	"encoding/json"

	chefcrypto "github.com/bhoriuchi/go-chef-crypto"
	"github.com/go-chef/chef"
)

// Decrypt decrypts an encrypted Chef Vault data bag item and returns the plaintext values.
func Decrypt(data chef.DataBagItem, key []byte) (chef.DataBagItem, error) {
	itemMap, err := DataBagItemMap(data)
	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{})
	for dbi, val := range itemMap {
		if dbi == "id" {
			out[dbi] = val
			continue
		}

		raw, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}

		var d interface{}
		if err := chefcrypto.Decrypt(key, raw, &d); err != nil {
			return nil, err
		}
		out[dbi] = d
	}
	return out, nil
}

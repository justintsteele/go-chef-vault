package item

import (
	"fmt"

	chefcrypto "github.com/bhoriuchi/go-chef-crypto"
	"github.com/go-chef/chef"
)

// EncryptedValue is the canonical Chef encrypted data bag item (v3).
// This is a type alias to avoid duplication and drift.
type EncryptedValue = chefcrypto.EncryptedDataBagItemV3

// VaultItem represents the encrypted data portion of a vault item.
type VaultItem struct {
	Id    string                    `json:"id"`
	Items map[string]EncryptedValue `json:"-"`
}

// DataBagItemMap converts a Chef DataBagItem into a map for processing decrypted vault content.
func DataBagItemMap(rawItem chef.DataBagItem) (map[string]interface{}, error) {
	if rawItem == nil {
		return nil, fmt.Errorf("nil DataBagItem")
	}

	m, ok := rawItem.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected DataBagItem type: %T", rawItem)
	}

	return m, nil
}

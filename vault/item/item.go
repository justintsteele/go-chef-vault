package item

import chefcrypto "github.com/bhoriuchi/go-chef-crypto"

// EncryptedValue is the canonical Chef encrypted data bag item (v3).
// This is a type alias to avoid duplication and drift.
type EncryptedValue = chefcrypto.EncryptedDataBagItemV3

// VaultItem represents the encrypted data portion of a vault item.
type VaultItem struct {
	Id        string                    `json:"id"`
	Items     map[string]EncryptedValue `json:"-"`
	decryptor VaultItemDecryptor
}

// VaultItemDecryptor defines the function used to decrypt vault item data.
type VaultItemDecryptor func(
	v *VaultItem,
	aesKey []byte,
) (map[string]interface{}, error)

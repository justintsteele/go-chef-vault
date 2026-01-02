package item

import chefcrypto "github.com/bhoriuchi/go-chef-crypto"

// EncryptedValue is the canonical Chef encrypted data bag item (v3).
// This is a type alias to avoid duplication and drift.
type EncryptedValue = chefcrypto.EncryptedDataBagItemV3

type VaultItem struct {
	Id        string                    `json:"id"`
	Items     map[string]EncryptedValue `json:"-"`
	decryptor VaultItemDecryptor
}

type VaultItemDecryptor func(
	v *VaultItem,
	aesKey []byte,
) (map[string]interface{}, error)

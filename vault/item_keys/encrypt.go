package item_keys

import "github.com/go-chef/chef"

var DefaultVaultItemKeyEncrypt VaultItemKeyEncryptor = (*VaultItemKeys).encryptKeys

// Encrypt lazy loader that encrypts the keys
func (k *VaultItemKeys) Encrypt(actors map[string]chef.AccessKey, secret []byte, out map[string]string) error {
	if k.encryptor == nil {
		k.encryptor = DefaultVaultItemKeyEncrypt
	}
	return k.encryptor(k, actors, secret, out)
}

// encryptKeys encrypts the public key of each actor in the vault
func (k *VaultItemKeys) encryptKeys(actors map[string]chef.AccessKey, secret []byte, out map[string]string) error {
	for actor, key := range actors {
		sharedSecret, err := encryptSharedSecret(key.PublicKey, secret)
		if err != nil {
			return err
		}
		out[actor] = sharedSecret
	}
	return nil
}

package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"

	"github.com/bhoriuchi/go-chef-crypto"
	"github.com/go-chef/chef"
)

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

var defaultVaultItemDecrypt VaultItemDecryptor = (*VaultItem).decryptItems

// decrypt lazy loader that decrypts the secret
func (i *VaultItem) decrypt(aesKey []byte) (map[string]interface{}, error) {
	if i.decryptor == nil {
		i.decryptor = defaultVaultItemDecrypt
	}
	return i.decryptor(i, aesKey)
}

// encryptContents creates the data half of the vault by creating an encrypted data bag item out of the Content in the payload encrypted the same shared secret
func encryptContents(payload *VaultPayload, secret []byte) (chef.DataBagItem, error) {
	item := make(map[string]any)
	item["id"] = payload.VaultItemName

	for k, v := range payload.Content {
		plaintext, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		// as per https://github.com/chef/chef/blob/main/chef-config/lib/chef-config/config.rb#L828
		// default encryption version is 3
		encrypted, err := chefcrypto.Encrypt(secret, plaintext, 3)
		if err != nil {
			return nil, err
		}
		item[k] = encrypted
	}

	return item, nil
}

// Decrypt the secret for real
func (i *VaultItem) decryptItems(aesKey []byte) (map[string]interface{}, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{})
	for name, ev := range i.Items {
		ct, err := base64.StdEncoding.DecodeString(cleanB64(ev.EncryptedData))
		if err != nil {
			return nil, err
		}
		nonce, err := base64.StdEncoding.DecodeString(cleanB64(ev.IV))
		if err != nil {
			return nil, err
		}
		tag, err := base64.StdEncoding.DecodeString(cleanB64(ev.AuthTag))
		if err != nil {
			return nil, err
		}
		ct = append(append([]byte{}, ct...), tag...)
		plaintext, err := gcm.Open(nil, nonce, ct, nil)
		if err != nil {
			return nil, err
		}

		var val map[string]interface{}
		if err := json.Unmarshal(plaintext, &val); err != nil {
			return nil, err
		}

		if len(val) == 1 {
			if jsonWrapper, ok := val["json_wrapper"]; ok {
				out[name] = jsonWrapper
				continue
			}
		}
		out[name] = val
	}
	return out, nil
}

// UnmarshalJSON overlay for VaultItem that types the response from the encrypted data bag
func (i *VaultItem) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	i.Items = make(map[string]EncryptedValue)
	for k, v := range raw {
		if k == "id" {
			i.Id = v.(string)
		}

		obj, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var ev EncryptedValue
		ev.EncryptedData = obj["encrypted_data"].(string)
		ev.IV = obj["iv"].(string)
		ev.AuthTag = obj["auth_tag"].(string)
		ev.Cipher = obj["cipher"].(string)

		if n, ok := obj["version"].(float64); ok {
			ev.Version = int(n)
		}

		i.Items[k] = ev
	}
	return nil
}

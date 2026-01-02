package item

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"strings"
)

// DefaultVaultItemDecrypt is a function variable used to allow tests to stub vault item decryption.
var DefaultVaultItemDecrypt VaultItemDecryptor = (*VaultItem).decryptItems

// Decrypt decrypts the vault item data using the provided AES key,
// delegating to the configured decryptor to allow test overrides.
func (i *VaultItem) Decrypt(aesKey []byte) (map[string]interface{}, error) {
	if i.decryptor == nil {
		i.decryptor = DefaultVaultItemDecrypt
	}
	return i.decryptor(i, aesKey)
}

// decryptItems performs AES-GCM decryption of all encrypted values in the vault item.
func (i *VaultItem) decryptItems(aesKey []byte) (map[string]interface{}, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{}, len(i.Items))

	for name, ev := range i.Items {
		val, err := decryptValue(gcm, ev)
		if err != nil {
			return nil, err
		}
		out[name] = val
	}

	return out, nil
}

// decryptValue decrypts a single encrypted vault item value.
func decryptValue(gcm cipher.AEAD, ev EncryptedValue) (any, error) {
	ct, err := base64.StdEncoding.DecodeString(CleanB64(ev.EncryptedData))
	if err != nil {
		return nil, err
	}

	nonce, err := base64.StdEncoding.DecodeString(CleanB64(ev.IV))
	if err != nil {
		return nil, err
	}

	tag, err := base64.StdEncoding.DecodeString(CleanB64(ev.AuthTag))
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
			return jsonWrapper, nil
		}
	}

	return val, nil
}

// CleanB64 removes extraneous whitespace from a base64-encoded string.
func CleanB64(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, s)
}

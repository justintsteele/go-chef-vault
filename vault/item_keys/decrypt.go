package item_keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"go-chef-vault/vault/item"
)

// DeriveAESKey is a hook to allow tests to stub AES key derivation.
var DeriveAESKey = AESKey

func DeriveAESKeyForVault(actorKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	return DeriveAESKey(actorKey, privateKey)
}

func AESKey(encryptedKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	encKey, err := base64.StdEncoding.DecodeString(
		item.CleanB64(encryptedKey),
	)
	if err != nil {
		return nil, err
	}

	secret, err := rsa.DecryptPKCS1v15(
		rand.Reader,
		privateKey,
		encKey,
	)
	if err != nil {
		return nil, fmt.Errorf("error decrypting key: %w", err)
	}

	sum := sha256.Sum256(secret)
	aesKey := sum[:]

	if len(aesKey) != 32 {
		return nil, fmt.Errorf("invalid AES key length")
	}

	return aesKey, nil
}

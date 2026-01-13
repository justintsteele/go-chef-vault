package item_keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/justintsteele/go-chef-vault/item"
)

// DeriveAESKey derives the AES key used to decrypt vault item data,
// delegating to DeriveAESKey to allow test overrides.
func DeriveAESKey(actorKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	sharedSecret, err := DecryptSharedSecret(actorKey, privateKey)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(sharedSecret)
	aesKey := sum[:]

	return aesKey, nil
}

// DecryptSharedSecret derives the shared secret from a private key.
func DecryptSharedSecret(actorKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	encKey, err := base64.StdEncoding.DecodeString(
		item.CleanB64(actorKey),
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
		return nil, fmt.Errorf("error decrypting shared secret: %w", err)
	}

	return secret, nil
}

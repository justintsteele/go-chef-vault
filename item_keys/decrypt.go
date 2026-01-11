package item_keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/justintsteele/go-chef-vault/item"
)

// DeriveAESKey is a function variable used to allow tests to stub AES key derivation.
var DeriveAESKey = AESKey

// DecryptSharedSecret is a function variable used to allow tests to stub shared secret decryption.
var DecryptSharedSecret = DecryptShared

// DeriveAESKeyForVault derives the AES key used to decrypt vault item data,
// delegating to DeriveAESKey to allow test overrides.
func DeriveAESKeyForVault(actorKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	return DeriveAESKey(actorKey, privateKey)
}

func DecryptActorSharedSecret(actorKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	return DecryptSharedSecret(actorKey, privateKey)
}

// DecryptShared derives the shared secret from a private key.
func DecryptShared(encryptedKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
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
		return nil, fmt.Errorf("error decrypting shared secret: %w", err)
	}

	return secret, nil
}

// AESKey derives the AES key used to decrypt an actor's vault data.
func AESKey(encryptedKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	sharedSecret, err := DecryptSharedSecret(encryptedKey, privateKey)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(sharedSecret)
	aesKey := sum[:]

	if len(aesKey) != 32 {
		return nil, fmt.Errorf("invalid AES key length")
	}

	return aesKey, nil
}

package item_keys

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"strings"
)

// DeriveAESKey derives the AES key used to decrypt vault item data,
func DeriveAESKey(actorKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	return DecryptSharedSecret(actorKey, privateKey)
}

// DecryptSharedSecret derives the shared secret from a private key.
func DecryptSharedSecret(actorKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	encKey, err := base64.StdEncoding.DecodeString(
		cleanB64(actorKey),
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

// cleanB64 removes extraneous whitespace from a base64-encoded string.
func cleanB64(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, s)
}

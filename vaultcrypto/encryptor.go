package vaultcrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

// EncryptSharedSecret takes in the pem of the actor and the shared secret, merges them, and encrypts it
func EncryptSharedSecret(publicKeyPEM string, secret []byte) (string, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return "", errors.New("failed to parse PEM block containing the public key")
	}

	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	pub, ok := pubAny.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("failed to parse RSA public key")
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pub, secret)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// GenSecret creates the shared secret used in the encryption of the vault items
func GenSecret(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

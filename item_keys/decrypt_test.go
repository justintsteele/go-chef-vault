package item_keys

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeriveAESKey_Returns32ByteKey(t *testing.T) {
	actorKey, privateKey, _ := genActorKey(t)

	key, err := DeriveAESKey(actorKey, privateKey)
	require.NoError(t, err)
	require.Len(t, key, 32)
}

func TestDecryptSharedSecret_RoundTrip(t *testing.T) {
	actorKey, privateKey, secret := genActorKey(t)

	got, err := DecryptSharedSecret(actorKey, privateKey)
	require.NoError(t, err)
	require.Equal(t, secret, got)
}

func TestDecryptSharedSecret_InvalidKeyFails(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	_, err := DecryptSharedSecret("not-base64", privateKey)
	require.Error(t, err)
}

func genActorKey(t *testing.T) (string, *rsa.PrivateKey, []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	secret := []byte("super-secret")

	encrypted, err := rsa.EncryptPKCS1v15(
		rand.Reader,
		&privateKey.PublicKey,
		secret,
	)
	require.NoError(t, err)

	return base64.StdEncoding.EncodeToString(encrypted), privateKey, secret
}

package item

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVaultItemDecrypt_RoundTrip(t *testing.T) {
	gcm, key := genGCM(t)
	item := &VaultItem{
		Items: map[string]EncryptedValue{
			"foo": encryptValue(t, gcm, map[string]any{"bar": "baz"}),
			"num": encryptValue(t, gcm, map[string]any{"n": float64(42)}),
		},
	}

	out, err := item.Decrypt(key)
	require.NoError(t, err)

	require.Equal(t, map[string]interface{}{
		"bar": "baz",
	}, out["foo"])

	require.Equal(t, map[string]interface{}{
		"n": float64(42),
	}, out["num"])
}

func TestDecryptValue_UnwrapsJSONWrapper(t *testing.T) {
	gcm, _ := genGCM(t)
	ev := encryptValue(t, gcm, map[string]any{
		"json_wrapper": "secret-value",
	})

	val, err := decryptValue(gcm, ev)
	require.NoError(t, err)

	require.Equal(t, "secret-value", val)
}

func TestVaultItemDecrypt_InvalidAESKey(t *testing.T) {
	item := &VaultItem{}

	_, err := item.Decrypt([]byte("short"))
	require.Error(t, err)
}

func TestCleanB64_RemovesWhitespace(t *testing.T) {
	in := " a\nb\rc\t== "
	out := CleanB64(in)
	require.Equal(t, "abc==", out)
}

func genGCM(t *testing.T) (cipher.AEAD, []byte) {
	t.Helper()

	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	block, err := aes.NewCipher(key)
	require.NoError(t, err)

	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	return gcm, key
}
func encryptValue(t *testing.T, gcm cipher.AEAD, val any) EncryptedValue {
	t.Helper()

	plaintext, err := json.Marshal(val)
	require.NoError(t, err)

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	require.NoError(t, err)

	ct := gcm.Seal(nil, nonce, plaintext, nil)

	tagLen := gcm.Overhead()
	ciphertext := ct[:len(ct)-tagLen]
	tag := ct[len(ct)-tagLen:]

	return EncryptedValue{
		EncryptedData: base64.StdEncoding.EncodeToString(ciphertext),
		IV:            base64.StdEncoding.EncodeToString(nonce),
		AuthTag:       base64.StdEncoding.EncodeToString(tag),
	}
}

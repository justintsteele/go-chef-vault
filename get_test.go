package vault

import (
	"crypto/rsa"
	"testing"

	"github.com/go-chef/chef"
	"github.com/stretchr/testify/require"
)

func TestService_GetItem(t *testing.T) {
	setupStubs(t)
	var calls []string

	ops := getOps{
		deriveAESKey: func(string, *rsa.PrivateKey) ([]byte, error) {
			calls = append(calls, "deriveAESKey")
			return []byte("secret"), nil
		},
		decrypt: func(chef.DataBagItem, []byte) (chef.DataBagItem, error) {
			calls = append(calls, "decrypt")
			return map[string]interface{}{
				"foo": "foo-value-1",
				"bar": "bar-value-1",
			}, nil
		},
	}

	_, err := service.getItem("vault1", "secret1", ops)
	require.NoError(t, err)
	require.Equal(t, []string{"deriveAESKey", "decrypt"}, calls)

}

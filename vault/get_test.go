package vault

import (
	"fmt"
	"net/http"
	"testing"
)

func TestService_GetItem(t *testing.T) {
	setup()
	defer teardown()

	defaultVaultItemDecrypt = func(_ *VaultItem, _ []byte) (map[string]interface{}, error) {
		return map[string]interface{}{
			"foo": "fake-foo-value",
			"bar": "fake-bar-value",
		}, nil
	}
	defer func() {
		defaultVaultItemDecrypt = (*VaultItem).decryptItems
	}()

	mux.HandleFunc("/data/vault1/secret1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"id": "secret1",
			"foo": {
				"encrypted_data": "foo-value",
				"iv": "foo-iv",
				"auth_tag": "foo-auth-tag",
				"version": 3,
				"cipher": "aes-256-gcm"
			},
			"bar": {
				"encrypted_data": "bar-value",
				"iv": "bar-iv",
				"auth_tag": "bar-auth-tag",
				"version": 3,
				"cipher": "aes-256-gcm"
			}
		}`)
	})

	// /data/vault1/prod_keys
	mux.HandleFunc("/data/vault1/secret1_keys", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"id": "secret1_keys",
			"admins": ["pivotal", "tester"],
			"clients": ["localhost"],
			"search_query": [],
			"mode": "default",
			"pivotal": "pivotal-private-key-b64\n",
			"localhost": "localhost-private-key-b64\n",
			"tester": "tester-private-key-b64\n"
		}`)
	})

	service.authorize = func(string) ([]byte, error) {
		return []byte("fake-aes-key"), nil
	}

	_, err := service.GetItem("vault1", "secret1")
	if err != nil {
		t.Errorf("Vaults.GetItem returned error: %v", err)
	}
}

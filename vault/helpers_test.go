package vault

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-chef/chef"
)

func stubVaultItemKeyEncrypt(t *testing.T) func() {
	t.Helper()
	orig := defaultVaultItemKeyEncrypt
	defaultVaultItemKeyEncrypt = func(_ *VaultItemKeys, _ map[string]chef.AccessKey, _ []byte, out map[string]string) error {
		out["tester"] = "ENCRYPTED RSA KEY"
		return nil
	}
	// return teardown
	return func() {
		defaultVaultItemKeyEncrypt = orig
	}
}

func stubVaultItemKeyDecrypt(t *testing.T) func() {
	t.Helper()

	// stub decrypt
	origDecrypt := defaultVaultItemDecrypt
	defaultVaultItemDecrypt = func(_ *VaultItem, _ []byte) (map[string]interface{}, error) {
		return map[string]interface{}{
			"foo": "fake-foo-value",
			"bar": "fake-bar-value",
		}, nil
	}

	// auth stub
	origAuthorize := service.authorize
	service.authorize = func(string) ([]byte, error) {
		return []byte("fake-aes-key"), nil
	}

	// return teardown
	return func() {
		defaultVaultItemDecrypt = origDecrypt
		service.authorize = origAuthorize
	}
}

func stubMuxCreate(t *testing.T) {
	t.Helper()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/data":
			fmt.Fprintf(w, `{"uri": "http://testhost/data/vault1"}`)
		case "/data/vault1":
			fmt.Fprintf(w, `{"uri": "http://testhost/data/vault1/secret1"}`)
		case "/users/tester/keys/default":
			fmt.Fprintf(w, `{
             			        "name": "default",
								"public_key": "RSA KEY",
								"expiration_date": "infinity"
                         	}`)
		case "/users/pivotal/keys/default":
			fmt.Fprintf(w, `{
             			        "name": "pivotal",
								"public_key": "RSA KEY",
								"expiration_date": "infinity"
                         	}`)
		case "/clients/testhost/keys/default":
			fmt.Fprintf(w, `{
             			        "name": "testhost",
								"public_key": "RSA KEY",
								"expiration_date": "infinity"
                         	}`)
		}
	})
}

func stubMuxGetItem(t *testing.T) {
	t.Helper()

	// item payload
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

	// keys payload
	mux.HandleFunc("/data/vault1/secret1_keys", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"id": "secret1_keys",
			"admins": ["pivotal", "tester"],
			"clients": ["testhost"],
			"search_query": [],
			"mode": "default",
			"pivotal": "pivotal-private-key-b64\n",
			"testhost": "testhost-private-key-b64\n",
			"tester": "tester-private-key-b64\n"
		}`)
	})
}

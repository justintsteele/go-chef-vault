package vault

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/vault/item"
	"github.com/justintsteele/go-chef-vault/vault/item_keys"
)

func setupStubs(t *testing.T) {
	t.Helper()

	setup(t)
	t.Cleanup(teardown)

	stubDeriveAESKey(t)
	stubVaultItemKeyDecrypt(t)
	stubVaultItemKeyEncrypt(t)
	stubMuxGetItem(t)
	stubMuxCreate(t)
}

func stubVaultItemKeyEncrypt(t *testing.T) {
	t.Helper()
	orig := item_keys.DefaultVaultItemKeyEncrypt
	item_keys.DefaultVaultItemKeyEncrypt = func(_ *item_keys.VaultItemKeys, actors map[string]chef.AccessKey, _ []byte, out map[string]string) error {
		for actor, key := range actors {
			out[actor] = fmt.Sprintf("ENCRYPTED %s", key.PublicKey)
		}
		return nil
	}

	t.Cleanup(func() { item_keys.DefaultVaultItemKeyEncrypt = orig })
}

func stubDeriveAESKey(t *testing.T) {
	t.Helper()

	orig := item_keys.DeriveAESKey
	item_keys.DeriveAESKey = func(_ string, _ *rsa.PrivateKey) ([]byte, error) {
		return make([]byte, 32), nil
	}

	t.Cleanup(func() { item_keys.DeriveAESKey = orig })
}

func stubVaultItemKeyDecrypt(t *testing.T) {
	t.Helper()

	// stub decrypt
	origDecrypt := item.DefaultVaultItemDecrypt
	item.DefaultVaultItemDecrypt = func(_ *item.VaultItem, _ []byte) (map[string]interface{}, error) {
		return map[string]interface{}{
			"foo": "fake-foo-value",
			"bar": "fake-bar-value",
		}, nil
	}

	origAuthorize := service.authorize
	service.authorize = func(string) ([]byte, error) {
		return []byte("fake-aes-key"), nil
	}

	t.Cleanup(func() {
		item.DefaultVaultItemDecrypt = origDecrypt
		service.authorize = origAuthorize
	})
}

func stubMuxCreate(t *testing.T) {
	t.Helper()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/data":
			_, _ = fmt.Fprintf(w, `{"uri": "http://localhost/data/vault1"}`)
		case "/data/vault1":
			_, _ = fmt.Fprintf(w, `{"uri": "http://localhost/data/vault1/secret1"}`)
		case "/users/tester/keys/default":
			_, _ = fmt.Fprintf(w, `{
             			        "name": "default",
								"public_key": "RSA KEY",
								"expiration_date": "infinity"
                         	}`)
		case "/users/pivotal/keys/default":
			_, _ = fmt.Fprintf(w, `{
             			        "name": "pivotal",
								"public_key": "RSA KEY",
								"expiration_date": "infinity"
                         	}`)
		case "/clients/testhost":
			_, _ = fmt.Fprintf(w, `{
             			        "name": "testhost",
								"public_key": "RSA KEY",
								"expiration_date": "infinity"
                         	}`)
		case "/clients/testhost3":
			_, _ = fmt.Fprintf(w, `{
             			        "name": "testhost3",
								"public_key": "RSA KEY",
								"expiration_date": "infinity"
                         	}`)
		case "/clients/testhost4":
			_, _ = fmt.Fprintf(w, `{
             			        "name": "testhost4",
								"public_key": "RSA KEY",
								"expiration_date": "infinity"
                         	}`)
		case "/clients/testhost/keys/default":
			_, _ = fmt.Fprintf(w, `{
             			        "name": "testhost",
								"public_key": "RSA KEY",
								"expiration_date": "infinity"
                         	}`)
		case "/clients/testhost3/keys/default":
			_, _ = fmt.Fprintf(w, `{
             			        "name": "testhost3",
								"public_key": "RSA KEY",
								"expiration_date": "infinity"
                         	}`)
		case "/clients/testhost4/keys/default":
			_, _ = fmt.Fprintf(w, `{
             			        "name": "testhost4",
								"public_key": "RSA KEY",
								"expiration_date": "infinity"
                         	}`)
		case "/search/node":
			_, _ = fmt.Fprintf(w, `{
								"total": 3,
								"start": 0,
								"rows": [
									{
										"url": "http://localhost/nodes/testhost",
										"data": { "name": "testhost" }
									},
									{
										"url": "http://localhost/nodes/testhost3",
										"data": { "name": "testhost3" }
									},
									{
										"url": "http://localhost/nodes/testhost4",
										"data": { "name": "testhost4" }
									}
								]
							}`)
		default:
			http.NotFound(w, r)
		}
	})
}

func stubMuxGetItem(t *testing.T) {
	t.Helper()

	// item payload
	mux.HandleFunc("/data/vault1/secret1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
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
		_, _ = fmt.Fprint(w, `{
			"id": "secret1_keys",
			"admins": ["pivotal", "tester"],
			"clients": ["testhost"],
			"search_query": "name:testhost*",
			"mode": "default",
			"pivotal": "pivotal-private-key-b64\n",
			"testhost": "testhost-private-key-b64\n",
			"tester": "tester-private-key-b64\n"
		}`)
	})
}

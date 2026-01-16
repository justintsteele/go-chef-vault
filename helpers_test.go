package vault

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/go-chef/chef"
)

const userid = "tester"

var (
	mux     *http.ServeMux
	server  *httptest.Server
	client  *chef.Client
	service *Service
)

func setup(t *testing.T) {
	t.Helper()

	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	privateKey := generateTestKeyPair(t)

	var err error
	client, err = chef.NewClient(&chef.Config{
		Name:                  userid,
		Key:                   privateKey,
		BaseURL:               server.URL,
		AuthenticationVersion: "1.0",
	})
	if err != nil {
		t.Fatalf("failed to create chef client: %v", err)
	}

	service = NewService(client)
}

func teardown() {
	server.Close()
}

func generateTestKeyPair(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	privDER := x509.MarshalPKCS1PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	})

	return string(privPEM)
}

func setupStubs(t *testing.T) {
	t.Helper()

	setup(t)
	t.Cleanup(teardown)

	stubMuxGetItem(t)
	stubMuxCreate(t)
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

	// vault payload
	mux.HandleFunc("/data/vault1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{
			"secret1": "http://localhost/data/vault1/secret1",
			"secret1_keys": "http://localhost/data/vault1/secret1_keys"
		}`)
	})

	// vault item payload
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

	// vault keys payload
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

	// encrypted data bag payload
	mux.HandleFunc("/data/encrdata1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{ "plaintext1": "http://localhost/data/encrdata1/encritem1" }`)
	})

	// encrypted data bag item payload
	mux.HandleFunc("/data/encrdata1/encritem1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"id": "encritem1",
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

	// data bag payload
	mux.HandleFunc("/data/databag1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{ "plaintext1": "http://localhost/data/databag1/plaintext1" }`)
	})

	// plaintext data bag payload
	mux.HandleFunc("/data/databag1/plaintext1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"id": "plaintext1",
			"plain": "plain-value"
		}`)
	})

}

func equalLists(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	a = slices.Clone(a)
	b = slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

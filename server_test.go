package vault

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
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

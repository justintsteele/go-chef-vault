package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-chef/chef"
)

func TestVaultsService_Update(t *testing.T) {
	setup()
	defer teardown()

	orig := defaultVaultItemKeyEncrypt
	defaultVaultItemKeyEncrypt = func(_ *VaultItemKeys, _ map[string]chef.AccessKey, _ []byte, out map[string]string) error {
		out["tester"] = "ENCRYPTED RSA KEY"
		return nil
	}
	defer func() {
		defaultVaultItemKeyEncrypt = orig
	}()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/data":
			fmt.Fprintf(w, `{"uri": "http://localhost/data/vault1"}`)
		case "/data/vault1":
			fmt.Fprintf(w, `{"uri": "http://localhost/data/vault1/secret"}`)
		case "/users/tester/keys/default":
			fmt.Fprintf(w, `{
             			        "name": "default",
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

	var raw map[string]interface{}
	vaultItem := `{"baz": "baz-value-1", "fuz": "fuz-value-2"}`
	if err := json.Unmarshal([]byte(vaultItem), &raw); err != nil {
		panic(err)
	}

	var admins []string
	admins = append(admins, client.Auth.ClientName)

	var clients []string
	clients = append(clients, "testhost")

	keysMode := KeysModeDefault
	pl := &VaultPayload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		Content:       raw,
		KeysMode:      &keysMode,
		SearchQuery:   nil,
		Admins:        admins,
		Clients:       clients,
	}

	response, err := service.Update(pl)
	if err != nil {
		t.Errorf("Vaults.Create returned error: %v", err)
	}

	want := &UpdateResponse{
		VaultResponse: VaultResponse{
			URI: fmt.Sprintf("%s/data/vault1", server.URL),
		},
		Data: &UpdateDataResponse{
			URI: fmt.Sprintf("%s/data/vault1/secret1", server.URL),
		},
		KeysURIs: []string{fmt.Sprintf("%s/data/vault1/secret1_keys", server.URL)},
	}

	if !reflect.DeepEqual(response, want) {
		t.Errorf("Vaults.Update returned %+v, want %+v", response, want)
	}
}

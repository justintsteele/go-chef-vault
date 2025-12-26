package vault

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-chef/chef"
)

func TestService_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/data":
			fmt.Fprint(w, `{"vault1":"http://testhost/data/vault1", "vault2":"http://testhost/data/vault2", "databag1":"http://testhost/data/databag1"}`)
		case "/data/vault1":
			fmt.Fprint(w, `{"secret1":"http://testhost/data/vault1/secret1", "secret1_keys":"http://testhost/data/vault1/secret1_keys"}`)
		case "/data/vault2":
			fmt.Fprint(w, `{"secret2":"http://testhost/data/vault2/secret2", "secret2_keys":"http://testhost/data/vault2/secret2_keys"}`)
		case "/data/databag1":
			fmt.Fprint(w, `{"foo":"http://testhost/data/databag1/foo"}`)
		default:
			http.NotFound(w, r)
		}
	})

	vaults, err := service.List()
	if err != nil {
		t.Errorf("Vaults.List returned error: %v", err)
	}
	want := &chef.DataBagListResult{"vault1": "http://testhost/data/vault1", "vault2": "http://testhost/data/vault2"}
	if !reflect.DeepEqual(vaults, want) {
		t.Errorf("Vaults.List returned %+v, want %+v", vaults, want)
	}
}

func TestService_ListItems(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/data/vault1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
						"secret1":"http://testhost/data/vault1/secret1", 
						"secret1_keys":"http://testhost/data/vault1/secret1_keys",
						"secret2":"http://testhost/data/vault2/secret2",
						"secret2_keys":"http://testhost/data/vault2/secret2_keys"
						}`)
	})
	vaults, err := service.ListItems("vault1")
	if err != nil {
		t.Errorf("Vaults.ListItems returned error: %v", err)
	}
	want := &chef.DataBagListResult{"secret1": "http://testhost/data/vault1/secret1", "secret2": "http://testhost/data/vault2/secret2"}
	if !reflect.DeepEqual(vaults, want) {
		t.Errorf("Vaults.ListItems returned %+v, want %+v", vaults, want)
	}
}

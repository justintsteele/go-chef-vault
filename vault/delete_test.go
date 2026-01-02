package vault

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestVaultsService_Delete(t *testing.T) {
	setup(t)
	t.Cleanup(teardown)

	mux.HandleFunc("/data/vault1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"name": "vault1", "json_class": "Chef::DataBag", "chef_type": "data_bag"}`)
	})

	response, err := service.Delete("vault1")
	if err != nil {
		t.Errorf("Vaults.Delete returned error: %v", err)
	}

	want := &DeleteResponse{
		Response: Response{
			URI: service.vaultURL("vault1"),
		},
	}

	if !reflect.DeepEqual(response, want) {
		t.Errorf("Vaults.Delete returned %+v, want %+v", response, want)
	}
}

func TestVaultsService_DeleteItem(t *testing.T) {
	setup(t)
	t.Cleanup(teardown)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/data/vault1/secret1":
			fmt.Fprintf(w, ``)
		case "/data/vault1/secret1_keys":
			fmt.Fprintf(w, ``)
		default:
			http.NotFound(w, r)
		}
	})
	response, err := service.DeleteItem("vault1", "secret1")
	if err != nil {
		t.Errorf("Vaults.Delete returned error: %v", err)
	}
	want := &DeleteResponse{
		Response: Response{
			URI: fmt.Sprintf("%s/%s", service.vaultURL("vault1"), "secret1"),
		},
		KeysURIs: []string{fmt.Sprintf("%s/%s", service.vaultURL("vault1"), "secret1_keys")},
	}
	if !reflect.DeepEqual(response, want) {
		t.Errorf("Vaults.DeleteItem returned %+v, want %+v", response, want)
	}
}

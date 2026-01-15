package vault

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestService_Delete(t *testing.T) {
	setup(t)
	t.Cleanup(teardown)

	mux.HandleFunc("/data/vault1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{"name": "vault1", "json_class": "Chef::DataBag", "chef_type": "data_bag"}`)
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

func TestService_DeleteItem(t *testing.T) {
	setupStubs(t)

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

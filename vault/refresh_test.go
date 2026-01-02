package vault

import (
	"fmt"
	"reflect"
	"testing"
)

func TestVaultService_RefreshClean(t *testing.T) {
	setupStubs(t)

	pl := &Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		Clean:         true,
	}

	got, err := service.Refresh(pl)
	if err != nil {
		t.Fatal(err)
	}

	want := &RefreshResponse{
		Response: Response{
			URI: fmt.Sprintf("%s/data/vault1", server.URL),
		},
		KeysURIs: []string{fmt.Sprintf("%s/data/vault1/secret1_keys", server.URL)},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func TestVaultService_RefreshSkip(t *testing.T) {
	setupStubs(t)

	pl := &Payload{
		VaultName:     "vault1",
		VaultItemName: "secret1",
		Clean:         false,
		SkipReencrypt: true,
	}

	got, err := service.Refresh(pl)
	if err != nil {
		t.Fatal(err)
	}

	want := &RefreshResponse{
		Response: Response{
			URI: fmt.Sprintf("%s/data/vault1", server.URL),
		},
		KeysURIs: []string{fmt.Sprintf("%s/data/vault1/secret1_keys", server.URL)},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

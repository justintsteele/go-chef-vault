package vault

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRefresh_Clean(t *testing.T) {
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

func TestRefresh_SkipReencrypt(t *testing.T) {
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

func TestRefresh_CleanClients(t *testing.T) {
	setupStubs(t)

	clients := []string{
		"testhost",
		"testhost2",
		"testhost3",
		"testhost4",
		"testhost5",
	}
	kept, removed, err := cleanClients(clients, service.clientExists)
	if err != nil {
		t.Fatal(err)
	}

	wantKept := []string{"testhost", "testhost3", "testhost4"}
	wantRemoved := []string{"testhost2", "testhost5"}
	if !reflect.DeepEqual(kept, wantKept) {
		t.Errorf("kept: %v, wantKept: %v", kept, wantKept)
	}
	if !reflect.DeepEqual(removed, wantRemoved) {
		t.Errorf("removed: %v, wantRemoved: %v", removed, wantRemoved)
	}
}

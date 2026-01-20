package item_keys

import (
	"slices"
	"testing"

	"github.com/go-chef/chef"
	"github.com/stretchr/testify/require"
)

func TestVaultItemKeys_BuildKeysItem(t *testing.T) {
	vik := &VaultItemKeys{
		Id:          "secret1_keys",
		Admins:      []string{"admin1"},
		SearchQuery: nil,
		Mode:        KeysModeDefault,
		Keys:        map[string]string{"client1": "some encrypted key"},
	}

	keys := vik.BuildKeysItem([]string{"client1"})
	require.Equal(t, keys["id"], "secret1_keys")
	require.NotNil(t, keys["client1"])
}

func TestMapKeys_SortsAndDedupes(t *testing.T) {
	input := map[string]chef.AccessKey{
		"testhost3": {},
		"tester":    {},
		"testhost":  {},
	}

	got := MapKeys(input)

	want := []string{"tester", "testhost", "testhost3"}

	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Fatalf("missing key %q in %v", w, got)
		}
	}
}

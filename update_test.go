package vault

import (
	"testing"

	"github.com/justintsteele/go-chef-vault/item_keys"
	"github.com/stretchr/testify/require"
)

func TestUpdate_ResolveUpdateContent(t *testing.T) {
	rawCurrent := map[string]interface{}{
		"foo": "fake-foo-value",
	}
	rawUpdate := map[string]interface{}{
		"bar": "fake-bar-value-2",
	}
	rawMerged := map[string]interface{}{
		"foo": "fake-foo-value",
		"bar": "fake-bar-value-2",
	}

	content, err := resolveContent(rawCurrent, rawUpdate)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, rawMerged, content)
}

func TestPayload_mergeKeyActors_MergesAdminsAndClients(t *testing.T) {
	setupStubs(t)
	payload, err := stubPayload([]string{"tester"}, []string{"testhost"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	keyState, err := service.loadKeysCurrentState(payload)
	if err != nil {
		t.Fatal(err)
	}

	updPayload, err := stubPayload([]string{"tester", "pivotal"}, []string{"testhost", "testhost3"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	updPayload.mergeKeyActors(keyState)

	got := keyState.Admins
	want := []string{"tester", "pivotal"}

	if !item_keys.EqualLists(got, want) {
		t.Errorf("All actors = %v, want %v", got, want)
	}
}

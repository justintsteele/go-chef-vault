package item_keys

import (
	"fmt"
	"slices"
	"sort"

	"github.com/go-chef/chef"
)

const (
	KeysModeDefault KeysMode = "default"
	KeysModeSparse  KeysMode = "sparse"
)

type KeysMode string

type KeysModeState struct {
	Current KeysMode `json:"current"`
	Desired KeysMode `json:"desired"`
}

type VaultItemKeys struct {
	Id          string            `json:"id"`
	Admins      []string          `json:"admins"`
	Clients     []string          `json:"clients"`
	SearchQuery interface{}       `json:"search_query"`
	Mode        KeysMode          `json:"mode"`
	Keys        map[string]string `json:"-"`
	encryptor   VaultItemKeyEncryptor
}

type VaultItemKeyEncryptor func(
	v *VaultItemKeys,
	actors map[string]chef.AccessKey,
	secret []byte,
	out map[string]string,
) error

type VaultItemKeysResult struct {
	URIs []string `json:"uris"`
}

// BuildKeysItem produces the map that will be used to create the keys data bag item
func (k *VaultItemKeys) BuildKeysItem() map[string]any {
	item := map[string]any{
		"id":           k.Id,
		"admins":       k.Admins,
		"clients":      k.Clients,
		"search_query": k.SearchQuery,
	}

	for actor, cipher := range k.Keys {
		item[actor] = cipher
	}

	return item
}

// MapKeys returns a slice of the keys of a map
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// MergeClients merges a payload client list and any clients returned from a search ensuring there are no duplicates
func MergeClients(a []string, b []string) []string {
	seen := make(map[string]struct{})

	for _, c := range a {
		seen[c] = struct{}{}
	}
	for _, c := range b {
		seen[c] = struct{}{}
	}

	out := make([]string, 0, len(seen))
	for c := range seen {
		out = append(out, c)
	}

	sort.Strings(out) // optional but nice
	return out
}

// EqualLists compares two lists to determine equality
func EqualLists(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	a = slices.Clone(a)
	b = slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

// DataBagItemMap returns a mapped data bag items for use in piecing together the decrypted secret
func DataBagItemMap(rawItem chef.DataBagItem) (map[string]interface{}, error) {
	if rawItem == nil {
		return nil, fmt.Errorf("nil DataBagItem")
	}

	m, ok := rawItem.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected DataBagItem type: %T", rawItem)
	}

	return m, nil
}

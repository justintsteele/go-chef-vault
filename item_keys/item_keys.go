package item_keys

import (
	"slices"
)

// KeysMode defines how vault access keys are managed during create and update.
type KeysMode string

const (
	// KeysModeDefault preserves a single shared key set for all authorized actors.
	KeysModeDefault KeysMode = "default"

	// KeysModeSparse assigns each authorized actor an independent key entry.
	KeysModeSparse KeysMode = "sparse"
)

// KeysModeState describes the current and desired KeysMode during a vault operation.
type KeysModeState struct {
	Current KeysMode `json:"current"`
	Desired KeysMode `json:"desired"`
}

// VaultItemKeys represents the key metadata for a vault item, including authorized actors,
// key storage mode, and an optional search query used to dynamically adjust client access.
type VaultItemKeys struct {
	Id          string            `json:"id"`
	Admins      []string          `json:"admins"`
	Clients     []string          `json:"clients"`
	SearchQuery interface{}       `json:"search_query"`
	Mode        KeysMode          `json:"mode"`
	Keys        map[string]string `json:"-"`
}

// VaultItemKeysResult represents the response returned from key-related vault operations.
type VaultItemKeysResult struct {
	URIs []string `json:"uris"`
}

// BuildKeysItem returns the data bag item used to persist vault keys.
func (k *VaultItemKeys) BuildKeysItem(id string, clients []string) map[string]any {
	item := map[string]any{
		"id":           id,
		"admins":       k.Admins,
		"clients":      clients,
		"search_query": k.SearchQuery,
		"mode":         k.Mode,
	}

	for actor, cipher := range k.Keys {
		item[actor] = cipher
	}

	return item
}

// MapKeys returns the keys of a map as a slice.
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// MergeClients merges two client lists, removing duplicates.
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

	return out
}

// DiffLists returns the elements in a that are not present in b (set difference: a - b).
func DiffLists(a, b []string) []string {
	bset := make(map[string]struct{}, len(b))
	for _, v := range b {
		bset[v] = struct{}{}
	}

	out := make([]string, 0)
	for _, v := range a {
		if _, ok := bset[v]; !ok {
			out = append(out, v)
		}
	}
	return out
}

// EqualLists reports whether two string slices contain the same elements, ignoring order.
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

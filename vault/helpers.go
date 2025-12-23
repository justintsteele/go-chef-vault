package vault

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-chef/chef"
)

// mergeClients merges a payload client list and any clients returned from a search ensuring there are no duplicates
func mergeClients(a, b []string) []string {
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

// Ensures base64 strings do not have extraneous characters in them
func cleanB64(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, s)
}

// dataBagItemMap returns a mapped data bag items for use in piecing together the decrypted secret
func dataBagItemMap(rawItem chef.DataBagItem) (map[string]interface{}, error) {
	if rawItem == nil {
		return nil, fmt.Errorf("nil DataBagItem")
	}

	m, ok := rawItem.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected DataBagItem type: %T", rawItem)
	}

	return m, nil
}

func toStringSlice(v interface{}) []string {
	raw, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, len(raw))
	for i := range raw {
		s, ok := raw[i].(string)
		if !ok {
			return nil
		}
		out[i] = s
	}
	return out
}

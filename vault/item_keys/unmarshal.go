package item_keys

import "encoding/json"

// UnmarshalJSON overlay for VaultItemKeys that types the response from the *_keys data bag
func (k *VaultItemKeys) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	k.Keys = make(map[string]string)

	for key, val := range raw {
		switch key {
		case "id":
			if s, ok := val.(string); ok {
				k.Id = s
			}
		case "admins":
			k.Admins = toStringSlice(val)
		case "clients":
			k.Clients = toStringSlice(val)
		case "search_query":
			// Preserve Chef semantics exactly: string OR [] OR nil
			k.SearchQuery = val
		case "mode":
			if s, ok := val.(string); ok {
				k.Mode = KeysMode(s)
			}
		default:
			// encrypted actor keys
			if s, ok := val.(string); ok {
				k.Keys[key] = s
			}
		}
	}

	return nil
}

// toStringSlice converts an interface from a databag to a slice of strings
func toStringSlice(v any) []string {
	switch t := v.(type) {
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, raw := range t {
			s, ok := raw.(string)
			if !ok {
				return nil
			}
			out = append(out, s)
		}
		return out
	case []string:
		return t
	case string:
		return []string{t}
	default:
		return nil
	}
}

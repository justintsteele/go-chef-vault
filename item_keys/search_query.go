package item_keys

import "fmt"

// NormalizeSearchQuery converts a schemaless search query value into a typed string or nil.
func NormalizeSearchQuery(v any) *string {
	if v == nil {
		return nil
	}

	switch q := v.(type) {
	case *string:
		if q == nil || *q == "" {
			return nil
		}
		return q

	case string:
		if q == "" {
			return nil
		}
		return &q

	default:
		s := fmt.Sprint(v)
		if s == "" || s == "<nil>" {
			return nil
		}
		return &s
	}
}

// ResolveSearchQuery applies Chef-Vault precedence rules for search_query values.
// If a search_query is not provided, Chef-Vault stores it as an empty array;
// otherwise it is stored as a string.
func ResolveSearchQuery(keyState interface{}, request *string) *string {
	if request != nil {
		return request
	}

	if ks, ok := keyState.(string); ok {
		return &ks
	}

	return nil
}

// EffectiveSearchQuery converts a normalized search query into the form expected by Chef-Vault,
// matching the behavior of ChefVault::ItemKeys initialization.
func EffectiveSearchQuery(q *string) interface{} {
	if q == nil {
		return []string{}
	}
	return *q
}

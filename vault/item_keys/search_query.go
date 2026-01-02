package item_keys

import "fmt"

// search_query handling is split across three layers:
//  1. NormalizeSearchQuery: schemaless → typed
//  2. ResolveSearchQuery: request vs existing precedence
//  3. EffectiveSearchQuery: Chef Vault storage semantics

// NormalizeSearchQuery helps with normalizing the search_query
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

// ResolveSearchQuery accounts for the behavior in Chef-Vault where if a search_query is not provided, it stores it as an empty array, all others are stored as a string
func ResolveSearchQuery(keyState interface{}, request *string) *string {
	if request != nil {
		return request
	}

	if ks, ok := keyState.(string); ok {
		return &ks
	}

	return nil
}

// EffectiveSearchQuery mimics behavior of ChefVault::ItemKeys initializer for search_query
func EffectiveSearchQuery(q *string) interface{} {
	if q == nil {
		return []string{}
	}
	return *q
}

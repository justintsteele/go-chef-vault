package item_keys

// ClientSearchPlan represents the payload sent to the Chef API search endpoint for node queries.
type ClientSearchPlan struct {
	Index  string
	Query  string
	Fields map[string]interface{}
}

// BuildClientSearchPlan creates a structured node search query from a normalized search string.
// If the query is nil or empty, it returns nil.
func BuildClientSearchPlan(q *string) *ClientSearchPlan {
	if q == nil || *q == "" {
		return nil
	}

	return &ClientSearchPlan{
		Index: "node",
		Query: *q,
		Fields: map[string]interface{}{
			"name": []string{"name"},
		},
	}
}

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
		return nil
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

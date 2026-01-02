package item_keys

import "encoding/json"

type ClientSearchPlan struct {
	Index  string
	Query  string
	Fields map[string]interface{}
}

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

func ExtractClients(rows []json.RawMessage) ([]string, error) {
	type node struct {
		Name string `json:"name"`
	}

	out := make([]string, 0, len(rows))
	for _, r := range rows {
		var n node
		if err := json.Unmarshal(r, &n); err != nil {
			return nil, err
		}
		if n.Name != "" {
			out = append(out, n.Name)
		}
	}
	return out, nil
}

package item

import (
	"fmt"

	"github.com/go-chef/chef"
)

// DataBagItemMap converts a Chef DataBagItem into a map for processing decrypted vault content.
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

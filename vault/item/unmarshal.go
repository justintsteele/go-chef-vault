package item

import "encoding/json"

// UnmarshalJSON implements custom decoding for VaultItem from a Chef encrypted data bag item.
func (i *VaultItem) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	i.Items = make(map[string]EncryptedValue)
	for k, v := range raw {
		if k == "id" {
			// Chef guarantees id is a string for encrypted data bag items.
			i.Id = v.(string)
		}

		obj, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		var ev EncryptedValue
		ev.EncryptedData = obj["encrypted_data"].(string)
		ev.IV = obj["iv"].(string)
		ev.AuthTag = obj["auth_tag"].(string)
		ev.Cipher = obj["cipher"].(string)

		if n, ok := obj["version"].(float64); ok {
			ev.Version = int(n)
		}

		i.Items[k] = ev
	}
	return nil
}

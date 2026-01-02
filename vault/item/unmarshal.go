package item

import "encoding/json"

// UnmarshalJSON overlay for VaultItem that types the response from the encrypted data bag
func (i *VaultItem) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	i.Items = make(map[string]EncryptedValue)
	for k, v := range raw {
		if k == "id" {
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

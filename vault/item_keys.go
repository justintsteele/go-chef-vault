package vault

import (
	"encoding/json"
	"fmt"
	"go-chef-vault/crypto"

	"github.com/go-chef/chef"
)

type VaultItemKeys struct {
	Id          string            `json:"id"`
	Admins      []string          `json:"admins"`
	Clients     []string          `json:"clients"`
	SearchQuery []string          `json:"search_query"`
	Mode        string            `json:"mode"`
	Keys        map[string]string `json:"-"`
	encryptor   VaultItemKeyEncryptor
}

type VaultItemKeyEncryptor func(
	v *VaultItemKeys,
	actors map[string]chef.AccessKey,
	secret []byte,
	out map[string]string,
) error

type VaultItemKeysResult struct {
	URIs []string `json:"uris"`
}

var defaultVaultItemKeyEncrypt VaultItemKeyEncryptor = (*VaultItemKeys).encryptKeys

// encrypt lazy loader that encrypts the keys
func (k *VaultItemKeys) encrypt(actors map[string]chef.AccessKey, secret []byte, out map[string]string) error {
	if k.encryptor == nil {
		k.encryptor = defaultVaultItemKeyEncrypt
	}
	return k.encryptor(k, actors, secret, out)
}

// encryptKeys encrypts the public key of each actor in the vault
func (k *VaultItemKeys) encryptKeys(actors map[string]chef.AccessKey, secret []byte, out map[string]string) error {
	for actor, key := range actors {
		sharedSecret, err := crypto.EncryptSharedSecret(key.PublicKey, secret)
		if err != nil {
			return err
		}
		out[actor] = sharedSecret
	}
	return nil
}

func (v *Service) createKeysDataBag(payload *VaultPayload, secret []byte, action string) (*VaultItemKeysResult, error) {
	mode := payload.EffectiveKeysMode()
	keys, err := v.buildKeys(payload, secret)
	result := &VaultItemKeysResult{}
	if err != nil {
		return nil, err
	}
	keys["mode"] = &mode

	switch mode {
	case KeysModeDefault:
		switch action {
		case "create":
			if kDbErr := v.Client.DataBags.CreateItem(payload.VaultName, &keys); kDbErr != nil {
				return nil, kDbErr
			}
		case "update":
			if kDbErr := v.Client.DataBags.UpdateItem(payload.VaultName, payload.VaultItemName+"_keys", &keys); kDbErr != nil {
				return nil, kDbErr
			}
		}
		result.URIs = append(result.URIs, fmt.Sprintf("%s/%s", v.vaultURL(payload.VaultName), payload.VaultItemName+"_keys"))

	case KeysModeSparse:
		baseKeys := map[string]any{
			"id":      keys["id"],
			"admins":  keys["admins"],
			"clients": keys["clients"],
			"mode":    keys["mode"],
		}

		if sq, ok := keys["search_query"].([]string); ok && sq != nil {
			baseKeys["search_query"] = sq
		} else {
			baseKeys["search_query"] = []string{}
		}

		switch action {
		case "create":
			if kDbErr := v.Client.DataBags.CreateItem(payload.VaultName, &baseKeys); kDbErr != nil {
				return nil, kDbErr
			}
		case "update":
			if kDbErr := v.Client.DataBags.UpdateItem(payload.VaultName, baseKeys["id"].(string), &baseKeys); kDbErr != nil {
				return nil, kDbErr
			}
		}
		result.URIs = append(result.URIs, fmt.Sprintf("%s/%s", v.vaultURL(payload.VaultName), baseKeys["id"].(string)))

		for k, val := range keys {
			switch k {
			case "id", "admins", "clients", "search_query", "mode":
				continue
			}
			sparseId := fmt.Sprintf("%s_key_%s", payload.VaultItemName, k)
			sparseItem := map[string]interface{}{
				"id": sparseId,
			}
			sparseItem[k] = val
			switch action {
			case "create":
				if kDbErr := v.Client.DataBags.CreateItem(payload.VaultName, &sparseItem); kDbErr != nil {
					return nil, kDbErr
				}
			case "update":
				_, err := v.Client.DataBags.GetItem(payload.VaultName, sparseId)
				if err == nil {
					if kDbErr := v.Client.DataBags.UpdateItem(payload.VaultName, sparseId, &sparseItem); kDbErr != nil {
						return nil, kDbErr
					}
				} else {
					if kDbErr := v.Client.DataBags.CreateItem(payload.VaultName, &sparseItem); kDbErr != nil {
						return nil, kDbErr
					}
				}
			}
			result.URIs = append(result.URIs, fmt.Sprintf("%s/%s", v.vaultURL(payload.VaultName), sparseId))
		}
	default:
		return result, fmt.Errorf("unsupported key format: %s", mode)
	}
	return result, nil
}

// buildKeys collects the public keys for all actors involved in the vault and encrypts them
func (v *Service) buildKeys(payload *VaultPayload, secret []byte) (map[string]any, error) {
	actors := make(map[string]chef.AccessKey)

	// admins
	for _, admin := range payload.Admins {
		key, err := v.Client.Users.GetKey(admin, "default")
		if err != nil {
			return nil, fmt.Errorf("fetch user %s failed: %w", admin, err)
		}
		actors[admin] = key
	}

	// explicit clients
	for _, client := range payload.Clients {
		key, err := v.Client.Clients.GetKey(client, "default")
		if err != nil {
			return nil, fmt.Errorf("fetch client %s failed: %w", client, err)
		}
		actors[client] = key
	}

	var searchQueries []string
	var searchedClients []string

	if payload.SearchQuery != nil {
		searchQueries = []string{*payload.SearchQuery}

		var err error
		searchedClients, err = v.getClientsFromSearch(payload)
		if err != nil {
			return nil, err
		}

		for _, client := range searchedClients {
			key, err := v.Client.Clients.GetKey(client, "default")
			if err != nil {
				return nil, fmt.Errorf("fetch client %s failed: %w", client, err)
			}
			actors[client] = key
		}
	}

	finalClients := mergeClients(payload.Clients, searchedClients)

	vik := &VaultItemKeys{
		Id:          payload.VaultItemName + "_keys",
		Admins:      payload.Admins,
		Clients:     finalClients,
		SearchQuery: searchQueries,
		Keys:        make(map[string]string),
	}

	if err := vik.encrypt(actors, secret, vik.Keys); err != nil {
		return nil, err
	}

	keysItem := map[string]any{
		"id":           vik.Id,
		"admins":       vik.Admins,
		"clients":      vik.Clients,
		"search_query": vik.SearchQuery,
	}

	for actor, cipher := range vik.Keys {
		keysItem[actor] = cipher
	}

	return keysItem, nil
}

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
			k.Id = val.(string)
		case "admins":
			k.Admins = toStringSlice(val)
		case "clients":
			k.Clients = toStringSlice(val)
		case "search_query":
			k.SearchQuery = toStringSlice(val)
		case "mode":
			k.Mode = val.(string)
		default:
			if s, ok := val.(string); ok {
				k.Keys[key] = s
			}
		}
	}
	return nil
}

package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-chef-vault/crypto"
	"net/http"

	"github.com/go-chef/chef"
)

type VaultItemKeys struct {
	Id          string            `json:"id"`
	Admins      []string          `json:"admins"`
	Clients     []string          `json:"clients"`
	SearchQuery interface{}       `json:"search_query"`
	Mode        KeysMode          `json:"mode"`
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

func (v *Service) buildDefaultKeys(payload *VaultPayload, keys *map[string]any, action string, out *VaultItemKeysResult) error {
	switch action {
	case "create":
		if err := v.Client.DataBags.CreateItem(payload.VaultName, &keys); err != nil {
			var chefErr *chef.ErrorResponse
			if errors.As(err, &chefErr) && chefErr.Response != nil && chefErr.Response.StatusCode == http.StatusConflict {
				if err := v.Client.DataBags.UpdateItem(payload.VaultName, payload.VaultItemName+"_keys", &keys); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	case "update":
		if err := v.Client.DataBags.UpdateItem(payload.VaultName, payload.VaultItemName+"_keys", &keys); err != nil {
			return err
		}
	}
	out.URIs = append(out.URIs, fmt.Sprintf("%s/%s", v.vaultURL(payload.VaultName), payload.VaultItemName+"_keys"))
	return nil
}

func (v *Service) buildSparseKeys(payload *VaultPayload, keys map[string]any, action string, out *VaultItemKeysResult) error {
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
		if err := v.Client.DataBags.CreateItem(payload.VaultName, &baseKeys); err != nil {
			return err
		}
	case "update":
		if err := v.Client.DataBags.UpdateItem(payload.VaultName, baseKeys["id"].(string), &baseKeys); err != nil {
			return err
		}
	}
	out.URIs = append(out.URIs, fmt.Sprintf("%s/%s", v.vaultURL(payload.VaultName), baseKeys["id"].(string)))

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
			if err := v.Client.DataBags.CreateItem(payload.VaultName, &sparseItem); err != nil {
				return err
			}
		case "update":
			if err := v.Client.DataBags.CreateItem(payload.VaultName, &sparseItem); err != nil {
				var chefErr *chef.ErrorResponse
				if errors.As(err, &chefErr) && chefErr.Response != nil && chefErr.Response.StatusCode == http.StatusConflict {
					if err := v.Client.DataBags.UpdateItem(payload.VaultName, sparseId, &sparseItem); err != nil {
						return err
					}
				} else {
					return err
				}
			}
		}
		out.URIs = append(out.URIs, fmt.Sprintf("%s/%s", v.vaultURL(payload.VaultName), sparseId))
	}
	return nil
}

func (v *Service) cleanupCurrentKeys(payload *VaultPayload, keysModeState *KeysModeState, keys map[string]any) error {
	switch keysModeState.Desired {
	case KeysModeDefault:
		// If Desired is "default", we need to clean up the sparse keys
		for key, _ := range keys {
			switch key {
			case "id", "admins", "clients", "search_query", "mode":
				continue
			}
			sparseId := fmt.Sprintf("%s_key_%s", payload.VaultItemName, key)
			if err := v.Client.DataBags.DeleteItem(payload.VaultName, sparseId); err != nil {
				return err
			}
		}
	case KeysModeSparse:
		// If Desired is "sparse", we need to clean up the base keys
		if err := v.Client.DataBags.DeleteItem(payload.VaultName, payload.VaultItemName+"_keys"); err != nil {
			return err
		}
	}
	return nil
}

func (v *Service) createKeysDataBag(payload *VaultPayload, keysModeState *KeysModeState, secret []byte, action string) (*VaultItemKeysResult, error) {
	mode := payload.EffectiveKeysMode()
	keys, err := v.buildKeys(payload, secret)
	result := &VaultItemKeysResult{}
	if err != nil {
		return nil, err
	}
	keys["mode"] = &mode

	if keysModeState.Current != keysModeState.Desired {
		if err := v.cleanupCurrentKeys(payload, keysModeState, keys); err != nil {
			return nil, err
		}
		action = "create"
	}

	switch mode {
	case KeysModeDefault:
		if err := v.buildDefaultKeys(payload, &keys, action, result); err != nil {
			return nil, err
		}
	case KeysModeSparse:
		if err := v.buildSparseKeys(payload, keys, action, result); err != nil {
			return nil, err
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

	var searchedClients []string

	if payload.SearchQuery != nil {
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
		SearchQuery: effectiveSearchQuery(payload.SearchQuery),
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

// effectiveSearchQuery mimics behavior of ChefVault::ItemKeys initializer for search_query
func effectiveSearchQuery(q *string) interface{} {
	if q == nil {
		return []string{}
	}
	return *q
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
			k.Mode = KeysMode(val.(string))
		default:
			if s, ok := val.(string); ok {
				k.Keys[key] = s
			}
		}
	}
	return nil
}

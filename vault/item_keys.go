package vault

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-chef-vault/vaultcrypto"
	"net/http"

	"github.com/go-chef/chef"
)

const (
	KeysModeDefault KeysMode = "default"
	KeysModeSparse  KeysMode = "sparse"
)

type KeysMode string

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
		sharedSecret, err := vaultcrypto.EncryptSharedSecret(key.PublicKey, secret)
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

func (v *Service) loadKeysCurrentState(payload *VaultPayload) (*VaultItemKeys, error) {

	raw, err := v.Client.DataBags.GetItem(payload.VaultName, payload.VaultItemName+"_keys")
	if err != nil {
		return nil, err
	}

	dim, err := dataBagItemMap(raw)
	if err != nil {
		return nil, err
	}

	vik := &VaultItemKeys{
		SearchQuery: dim["search_query"], // preserved as string OR []
	}

	if id, ok := dim["id"].(string); ok {
		vik.Id = id
	}

	if admins, ok := dim["admins"]; ok {
		vik.Admins = toStringSlice(admins)
	}

	if clients, ok := dim["clients"]; ok {
		vik.Clients = toStringSlice(clients)
	}

	if mode, ok := dim["mode"].(string); ok {
		vik.Mode = KeysMode(mode)
	}

	return vik, nil
}

func (v *Service) createKeysDataBag(payload *VaultPayload, keysModeState *KeysModeState, secret []byte, action string) (*VaultItemKeysResult, error) {
	mode := payload.effectiveKeysMode()
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

// buildKeys collects actor public keys and builds the encrypted vault keys item
func (v *Service) buildKeys(payload *VaultPayload, secret []byte) (map[string]any, error) {
	admins := make(map[string]chef.AccessKey)
	clients := make(map[string]chef.AccessKey)

	// Admins are required
	v.collectAdmins(payload.Admins, admins)

	if len(admins) == 0 {
		return nil, fmt.Errorf("none of the specified admins have public keys")
	}

	// Explicit clients
	v.collectClients(payload.Clients, clients)

	// Clients from search
	var searchedClients []string
	if payload.SearchQuery != nil {
		var err error
		searchedClients, err = v.getClientsFromSearch(payload)
		if err != nil {
			return nil, err
		}
		v.collectClients(searchedClients, clients)
	}

	finalClients := mapKeys(clients)

	vik := &VaultItemKeys{
		Id:          payload.VaultItemName + "_keys",
		Admins:      payload.Admins,
		Clients:     finalClients,
		SearchQuery: effectiveSearchQuery(payload.SearchQuery),
		Keys:        make(map[string]string),
	}

	// produce a list of all actors we need to encrypt
	actors := make(map[string]chef.AccessKey, len(admins)+len(clients))
	for k, v := range admins {
		actors[k] = v
	}
	for k, v := range clients {
		actors[k] = v
	}

	if err := vik.encrypt(actors, secret, vik.Keys); err != nil {
		return nil, err
	}

	return buildKeysItem(vik), nil
}

// collectAdmins collects the public keys for the given admins
func (v *Service) collectAdmins(names []string, admins map[string]chef.AccessKey) {
	for _, name := range names {
		key, err := v.Client.Users.GetKey(name, "default")
		if err != nil {
			fmt.Printf("admin %q has no public key, skipping: %v\n", name, err)
			continue
		}
		admins[name] = key
	}
}

// collectClients collects the public keys for the given clients
func (v *Service) collectClients(names []string, clients map[string]chef.AccessKey) {
	for _, name := range names {
		key, err := v.Client.Clients.GetKey(name, "default")
		if err != nil {
			fmt.Printf("client %q has no public key, skipping: %v\n", name, err)
			continue
		}
		clients[name] = key
	}
}

// buildKeysItem produces the map that will be used to create the keys data bag item
func buildKeysItem(vik *VaultItemKeys) map[string]any {
	item := map[string]any{
		"id":           vik.Id,
		"admins":       vik.Admins,
		"clients":      vik.Clients,
		"search_query": vik.SearchQuery,
	}

	for actor, cipher := range vik.Keys {
		item[actor] = cipher
	}

	return item
}

// effectiveKeysMode allows the payload to not contain a KeysMode and still enact the default
func (p *VaultPayload) effectiveKeysMode() KeysMode {
	if p.KeysMode == nil {
		return KeysModeDefault
	}
	return *p.KeysMode
}

// effectiveSearchQuery mimics behavior of ChefVault::ItemKeys initializer for search_query
func effectiveSearchQuery(q *string) interface{} {
	if q == nil {
		return []string{}
	}
	return *q
}

// resolveSearchQuery accounts for the behavior in Chef-Vault where if a search_query is not provided, it stores it as an empty array, all others are stored as a string
func resolveSearchQuery(keyState interface{}, request *string) *string {
	if request != nil {
		return request
	}

	if ks, ok := keyState.(string); ok {
		return &ks
	}

	return nil
}

// resolveKeysMode resolves the desired keys mode from the payload and the current mode
func resolveKeysMode(current KeysMode, payload *VaultPayload) (KeysMode, *KeysModeState) {
	if payload.KeysMode == nil {
		return current, &KeysModeState{
			Current: current,
			Desired: current,
		}
	}

	return *payload.KeysMode, &KeysModeState{
		Current: current,
		Desired: *payload.KeysMode,
	}
}

// mergeKeyActors merges the actors from the payload into the state, respecting the clean flag
func mergeKeyActors(state *VaultItemKeys, payload *VaultPayload) {
	if payload.Admins != nil {
		state.Admins = mergeClients(state.Admins, payload.Admins)
	}

	if payload.Clean {
		state.Clients = nil
	}

	if payload.Clients != nil {
		if payload.Clean {
			state.Clients = payload.Clients
		} else {
			state.Clients = mergeClients(state.Clients, payload.Clients)
		}
	}
}

// resolveUpdateContent merges the payload content with the current content, respecting the clean flag
func resolveUpdateContent(v *Service, payload *VaultPayload) (map[string]any, error) {
	current, err := v.GetItem(payload.VaultName, payload.VaultItemName)
	if err != nil {
		return nil, err
	}

	if payload.Content == nil {
		return dataBagItemMap(current)
	}

	merged := make(map[string]any)
	for k, v := range payload.Content {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		merged[k] = b
	}

	return dataBagItemMap(merged)
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

package vault

import (
	"encoding/json"
	"fmt"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/vault/cheferr"
	"github.com/justintsteele/go-chef-vault/vault/item_keys"
)

func (s *Service) buildDefaultKeys(payload *Payload, keys *map[string]any, out *item_keys.VaultItemKeysResult) error {
	if err := s.Client.DataBags.CreateItem(payload.VaultName, &keys); err != nil {
		if cheferr.IsConflict(err) {
			if err := s.Client.DataBags.UpdateItem(payload.VaultName, payload.VaultItemName+"_keys", &keys); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	out.URIs = append(out.URIs, fmt.Sprintf("%s/%s", s.vaultURL(payload.VaultName), payload.VaultItemName+"_keys"))
	return nil
}

func (s *Service) buildSparseKeys(payload *Payload, keys map[string]any, out *item_keys.VaultItemKeysResult) error {
	baseKeys := map[string]any{
		"id":           keys["id"],
		"admins":       keys["admins"],
		"clients":      keys["clients"],
		"mode":         keys["mode"],
		"search_query": keys["search_query"],
	}

	if err := s.Client.DataBags.CreateItem(payload.VaultName, &baseKeys); err != nil {
		if cheferr.IsConflict(err) {
			if err := s.Client.DataBags.UpdateItem(payload.VaultName, baseKeys["id"].(string), &baseKeys); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	out.URIs = append(out.URIs, fmt.Sprintf("%s/%s", s.vaultURL(payload.VaultName), baseKeys["id"].(string)))

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
		if err := s.Client.DataBags.CreateItem(payload.VaultName, &sparseItem); err != nil {
			if cheferr.IsConflict(err) {
				if err := s.Client.DataBags.UpdateItem(payload.VaultName, sparseId, &sparseItem); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		out.URIs = append(out.URIs, fmt.Sprintf("%s/%s", s.vaultURL(payload.VaultName), sparseId))
	}
	return nil
}

func (s *Service) cleanupCurrentKeys(payload *Payload, keysModeState *item_keys.KeysModeState, keys map[string]any) error {
	switch keysModeState.Desired {
	case item_keys.KeysModeDefault:
		// If Desired is "default", we need to clean up the sparse keys
		for key := range keys {
			switch key {
			case "id", "admins", "clients", "search_query", "mode":
				continue
			}
			sparseId := fmt.Sprintf("%s_key_%s", payload.VaultItemName, key)
			if err := s.Client.DataBags.DeleteItem(payload.VaultName, sparseId); err != nil {
				return err
			}
		}
	case item_keys.KeysModeSparse:
		// If Desired is "sparse", we need to clean up the base keys
		if err := s.Client.DataBags.DeleteItem(payload.VaultName, payload.VaultItemName+"_keys"); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) loadKeysCurrentState(payload *Payload) (*item_keys.VaultItemKeys, error) {
	raw, err := s.Client.DataBags.GetItem(
		payload.VaultName,
		payload.VaultItemName+"_keys",
	)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var vik item_keys.VaultItemKeys
	if err := json.Unmarshal(b, &vik); err != nil {
		return nil, err
	}

	return &vik, nil
}

func (s *Service) createKeysDataBag(payload *Payload, keysModeState *item_keys.KeysModeState, secret []byte) (*item_keys.VaultItemKeysResult, error) {
	mode := payload.effectiveKeysMode()
	keys, err := s.buildKeys(payload, secret)
	result := &item_keys.VaultItemKeysResult{}
	if err != nil {
		return nil, err
	}
	keys["mode"] = &mode

	if keysModeState.Current != keysModeState.Desired {
		if err := s.cleanupCurrentKeys(payload, keysModeState, keys); err != nil {
			return nil, err
		}
	}

	switch mode {
	case item_keys.KeysModeDefault:
		if err := s.buildDefaultKeys(payload, &keys, result); err != nil {
			return nil, err
		}
	case item_keys.KeysModeSparse:
		if err := s.buildSparseKeys(payload, keys, result); err != nil {
			return nil, err
		}
	default:
		return result, fmt.Errorf("unsupported key format: %s", mode)
	}
	return result, nil
}

// buildKeys collects actor public keys and builds the encrypted vault keys item
func (s *Service) buildKeys(payload *Payload, secret []byte) (map[string]any, error) {
	admins := make(map[string]chef.AccessKey)
	clients := make(map[string]chef.AccessKey)

	// Admins are required
	s.collectAdmins(payload.Admins, admins)

	if len(admins) == 0 {
		return nil, fmt.Errorf("none of the specified admins have public keys")
	}

	// Explicit clients
	s.collectClients(payload.Clients, clients)

	// Clients from search
	var searchedClients []string
	if payload.SearchQuery != nil {
		var err error
		searchedClients, err = s.getClientsFromSearch(payload)
		if err != nil {
			return nil, err
		}
		s.collectClients(searchedClients, clients)
	}

	finalClients := item_keys.MapKeys(clients)

	vik := &item_keys.VaultItemKeys{
		Id:          payload.VaultItemName + "_keys",
		Admins:      payload.Admins,
		Clients:     finalClients,
		SearchQuery: item_keys.EffectiveSearchQuery(payload.SearchQuery),
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

	if err := vik.Encrypt(actors, secret, vik.Keys); err != nil {
		return nil, err
	}

	return vik.BuildKeysItem(), nil
}

// collectAdmins collects the public keys for the given admins
func (s *Service) collectAdmins(names []string, admins map[string]chef.AccessKey) {
	for _, name := range names {
		key, err := s.Client.Users.GetKey(name, "default")
		if err != nil {
			fmt.Printf("admin %q has no public key, skipping: %s\n", name, err)
			continue
		}
		admins[name] = key
	}
}

// collectClients collects the public keys for the given clients
func (s *Service) collectClients(names []string, clients map[string]chef.AccessKey) {
	for _, name := range names {
		key, err := s.Client.Clients.GetKey(name, "default")
		if err != nil {
			fmt.Printf("client %q has no public key, skipping: %s\n", name, err)
			continue
		}
		clients[name] = key
	}
}

// resolveUpdateContent merges the payload content with the current content, respecting the clean flag
func (s *Service) resolveUpdateContent(p *Payload) (map[string]any, error) {
	current, err := s.GetItem(p.VaultName, p.VaultItemName)
	if err != nil {
		return nil, err
	}

	if p.Content == nil {
		return item_keys.DataBagItemMap(current)
	}

	merged := make(map[string]any, len(p.Content))
	for k, v := range p.Content {
		merged[k] = v
	}
	return merged, nil
}

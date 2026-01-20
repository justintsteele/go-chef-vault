package vault

import (
	"encoding/json"
	"fmt"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/cheferr"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

// loadKeysCurrentState retrieves the data from the default keys data bag item prior to actions being taken on the vault.
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

	return vik.BuildKeysItem(finalClients), nil
}

// createKeysDataBag prepares the item_keys.VaultItemKeysResult to be written out as data bag items.
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

	if err := s.writeKeys(payload, mode, keys, result); err != nil {
		return nil, err
	}

	return result, nil
}

// writeKeys creates the default and sparse keys data bag items as specified in the item_keys.VaultItemKeysResult.
func (s *Service) writeKeys(payload *Payload, mode item_keys.KeysMode, keys map[string]any, result *item_keys.VaultItemKeysResult) error {
	switch mode {
	case item_keys.KeysModeDefault:
		return s.writeDefaultKeys(payload, &keys, result)
	case item_keys.KeysModeSparse:
		return s.writeSparseKeys(payload, keys, result)
	default:
		return fmt.Errorf("unsupported key format: %s", mode)
	}
}

// writeDefaultKeys constructs and writes the default keys data bag item.
func (s *Service) writeDefaultKeys(payload *Payload, keys *map[string]any, out *item_keys.VaultItemKeysResult) error {
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

// writeSparseKeys constructs and writes the sparse keys data bag items.
func (s *Service) writeSparseKeys(payload *Payload, keys map[string]any, out *item_keys.VaultItemKeysResult) error {
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

// collectAdmins collects the public keys for the given admins.
func (s *Service) collectAdmins(names []string, admins map[string]chef.AccessKey) {
	for _, name := range names {
		key, err := s.Client.Users.GetKey(name, "default")
		if err != nil {
			// misses here should be non-fatal so that we continue to get the keys for the actors that exist.
			continue
		}
		admins[name] = key
	}
}

// collectClients collects the public keys for the given clients.
func (s *Service) collectClients(names []string, clients map[string]chef.AccessKey) {
	for _, name := range names {
		key, err := s.clientPublicKey(name)
		if err != nil {
			// misses here should be non-fatal so that we continue to get the keys for the actors that exist.
			continue
		}
		clients[name] = key
	}
}

// clientPublicKey retrieves the public key for a specified actor.
func (s *Service) clientPublicKey(actor string) (chef.AccessKey, error) {
	return s.Client.Clients.GetKey(actor, "default")
}

// cleanupCurrentKeys migrates keys between default and sparse keys modes.
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

// cleanUnknownClients removes non-existent clients and prunes their keys from keyState.
func (s *Service) cleanUnknownClients(payload *Payload, keyState *item_keys.VaultItemKeys, clients []string) (kept []string, removed []string, err error) {
	kept, removed, err = resolveClients(clients, s.clientExists)
	if err != nil {
		return nil, nil, err
	}
	if len(removed) != 0 {
		if err := s.pruneKeys(removed, keyState, payload); err != nil {
			return nil, nil, err
		}
	}
	return kept, removed, nil
}

// pruneKeys removes the keys for the requested actors.
func (s *Service) pruneKeys(actors []string, keyState *item_keys.VaultItemKeys, payload *Payload) error {
	for _, actor := range actors {
		keyState.PruneActor(actor)
		if keyState.Mode == item_keys.KeysModeSparse {
			if err := s.deleteSparseKeys(payload.VaultName, payload.VaultItemName, actor, &DeleteResponse{}); err != nil {
				return err
			}
		}
	}
	return nil
}

// deleteDefaultKeys removes the base keys and any actor keys stored in default mode.
func (s *Service) deleteDefaultKeys(name string, item string, out *DeleteResponse) error {
	itemKeysUri := fmt.Sprintf("%s/%s", s.vaultURL(name), item+"_keys")
	if err := s.Client.DataBags.DeleteItem(name, item+"_keys"); err != nil {
		return err
	}
	out.KeysURIs = append(out.KeysURIs, itemKeysUri)
	return nil
}

// deleteSparseKeys removes all actor keys and the base sparse keys item.
func (s *Service) deleteSparseKeys(name string, item string, actor interface{}, out *DeleteResponse) error {
	baseKeyId := fmt.Sprintf("%s_keys", item)
	baseUri := fmt.Sprintf("%s/%s", s.vaultURL(name), baseKeyId)
	out.KeysURIs = append(out.KeysURIs, baseUri)

	var actors []string
	switch actor := actor.(type) {
	case string:
		actors = []string{actor}
	case []string:
		actors = actor
	}

	for _, actor := range actors {
		sparseId := fmt.Sprintf("%s_key_%s", item, actor)
		adminKeyUri := fmt.Sprintf("%s/%s", s.vaultURL(name), sparseId)
		if err := s.Client.DataBags.DeleteItem(name, sparseId); err != nil {
			if !cheferr.IsNotFound(err) {
				return err
			}
		}
		out.KeysURIs = append(out.KeysURIs, adminKeyUri)
	}
	return nil
}

// clientExists performs a client lookup to validate the requested client still exists in the Chef Server.
func (s *Service) clientExists(name string) (bool, error) {
	_, err := s.Client.Clients.Get(name)
	if err == nil {
		return true, nil
	}

	if cheferr.IsNotFound(err) {
		return false, nil
	}

	return false, err
}

// resolveClients partitions clients into those that still exist on the Chef server and those that do not.
func resolveClients(clients []string, exists func(string) (bool, error)) (kept []string, removed []string, err error) {
	kept = clients[:0]

	for _, c := range clients {
		ok, err := exists(c)
		if err != nil {
			return nil, nil, err
		}

		if ok {
			kept = append(kept, c)
		} else {
			removed = append(removed, c)
		}
	}

	return kept, removed, nil
}

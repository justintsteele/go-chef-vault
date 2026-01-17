package vault

import (
	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

// RemoveResponse represents the structure of the response from a Remove operation.
type RemoveResponse = UpdateResponse

// RemoveDataResponse represents the response returned after removing data from the vault item.
type RemoveDataResponse struct {
	URI string `json:"uri"`
}

// removeOps defines the callable operations required to execute an Remove request.
type removeOps struct {
	getItem func(string, string) (chef.DataBagItem, error)
	update  func(*Payload) (*UpdateResponse, error)
}

// Remove removes clients, admins, or data keys from an existing vault item.
//
// References:
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_remove.rb
func (s *Service) Remove(payload *Payload) (*RemoveResponse, error) {
	ops := removeOps{
		getItem: s.GetItem,
		update:  s.Update,
	}
	return s.remove(payload, ops)
}

// remove is the worker called by the public API with the operational methods to complete the Remove request.
func (s *Service) remove(payload *Payload, ops removeOps) (*RemoveResponse, error) {
	keyState, err := s.loadKeysCurrentState(payload)
	if err != nil {
		return nil, err
	}

	finalPayload := &Payload{
		VaultName:     payload.VaultName,
		VaultItemName: payload.VaultItemName,
		Clients:       keyState.Clients,
		Admins:        keyState.Admins,
	}

	if payload.CleanUnknown {
		resolvedClients, _, err := s.cleanUnknownClients(payload, keyState, keyState.Clients)
		if err != nil {
			return nil, err
		}

		finalPayload.Clients = resolvedClients
	}

	if err := s.resolveActors(payload, keyState); err != nil {
		return nil, err
	}

	if payload.Content != nil {
		current, err := ops.getItem(payload.VaultName, payload.VaultItemName)
		if err != nil {
			return nil, err
		}

		dbi, err := item.DataBagItemMap(current)
		if err != nil {
			return nil, err
		}

		removeContent, ok := pruneData(dbi, payload.Content)
		if ok {
			finalPayload.Content = removeContent.(map[string]any)
		}
	}

	removed, err := ops.update(finalPayload)
	if err != nil {
		return nil, err
	}

	return removed, nil
}

// resolveActors removes actors and their keys.
func (s *Service) resolveActors(payload *Payload, keyState *item_keys.VaultItemKeys) error {
	toRemove := make([]string, 0)

	if payload.SearchQuery != nil {
		found, err := s.getClientsFromSearch(payload)
		if err != nil {
			return err
		}
		toRemove = append(toRemove, found...)
	}

	if payload.Clients != nil {
		toRemove = append(toRemove, payload.Clients...)
	}

	if payload.Admins != nil {
		toRemove = append(toRemove, payload.Admins...)
	}

	if len(toRemove) == 0 {
		return nil
	}

	if err := s.pruneKeys(toRemove, keyState, payload); err != nil {
		return err
	}
	return nil
}

// pruneData recursively removes keys from existing data based on the shape of the remove payload.
func pruneData(existing, remove any) (any, bool) {
	switch rem := remove.(type) {

	case map[string]any:
		exMap, ok := existing.(map[string]any)
		if !ok {
			return nil, false
		}

		out := make(map[string]any)
		for k, v := range exMap {
			rv, exists := rem[k]
			if !exists {
				out[k] = v
				continue
			}

			if pruned, keep := pruneData(v, rv); keep {
				out[k] = pruned
			}
		}

		if len(out) == 0 {
			return nil, false
		}
		return out, true

	default:
		return nil, false
	}
}

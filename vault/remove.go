package vault

import (
	"github.com/justintsteele/go-chef-vault/vault/item"
	"github.com/justintsteele/go-chef-vault/vault/item_keys"
)

// RemoveResponse represents the structure of the response from a Remove operation.
type RemoveResponse = UpdateResponse

// RemoveDataResponse represents the response returned after removing data from the vault item.
type RemoveDataResponse struct {
	URI string `json:"uri"`
}

// Remove removes clients, admins, or data keys from an existing vault item.
//
// References:
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_remove.rb
func (s *Service) Remove(payload *Payload) (*RemoveResponse, error) {
	keyState, err := s.loadKeysCurrentState(payload)
	if err != nil {
		return nil, err
	}

	newClients, err := s.resolveClients(payload, keyState)
	if err != nil {
		return nil, err
	}

	newAdmins := keyState.Admins
	if payload.Admins != nil {
		newAdmins = resolveActors(keyState.Admins, payload.Admins)
	}

	removePayload := &Payload{
		VaultName:     payload.VaultName,
		VaultItemName: payload.VaultItemName,
		Clients:       newClients,
		Admins:        newAdmins,
	}

	if payload.Clean {
		removePayload.Clean = payload.Clean
	}
	removePayload.mergeKeyActors(keyState)

	if payload.Content != nil {
		removeContent, err := s.resolveRemoveContent(payload)
		if err != nil {
			return nil, err
		}
		removePayload.Content = removeContent
	}

	removed, err := s.Update(removePayload)
	if err != nil {
		return nil, err
	}

	return removed, nil
}

// resolveClients returns the list of desired clients after the requested removals.
func (s *Service) resolveClients(payload *Payload, keyState *item_keys.VaultItemKeys) ([]string, error) {
	toRemove := make([]string, 0)

	if payload.SearchQuery != nil {
		found, err := s.getClientsFromSearch(payload)
		if err != nil {
			return nil, err
		}
		toRemove = append(toRemove, found...)
	}

	if payload.Clients != nil {
		toRemove = append(toRemove, payload.Clients...)
	}

	if len(toRemove) == 0 {
		return keyState.Clients, nil
	}

	return resolveActors(keyState.Clients, toRemove), nil
}

// resolveRemoveContent returns the content of the vault after the requested keys
// (and nested keys) have been removed.
func (s *Service) resolveRemoveContent(p *Payload) (map[string]any, error) {
	current, err := s.GetItem(p.VaultName, p.VaultItemName)
	if err != nil {
		return nil, err
	}
	dbi, err := item.DataBagItemMap(current)
	if err != nil {
		return nil, err
	}

	kept := make(map[string]any)

	for key, value := range dbi {
		if removeVal, ok := p.Content[key]; ok {
			if pruned, keep := pruneData(value, removeVal); keep {
				kept[key] = pruned
			}
		} else {
			kept[key] = value
		}
	}
	return kept, nil
}

// resolveActors returns the remaining actors after requested removals.
func resolveActors(have, remove []string) []string {
	if len(have) == 0 || len(remove) == 0 {
		return have
	}

	removeSet := make(map[string]struct{}, len(remove))
	for _, r := range remove {
		removeSet[r] = struct{}{}
	}

	out := make([]string, 0, len(have))
	for _, h := range have {
		if _, found := removeSet[h]; !found {
			out = append(out, h)
		}
	}

	return out
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

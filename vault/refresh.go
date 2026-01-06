package vault

import (
	"errors"
	"go-chef-vault/vault/cheferr"
	"go-chef-vault/vault/item_keys"
)

// RefreshResponse intentionally mirrors UpdateResponse for API parity.
type RefreshResponse = UpdateResponse

// Refresh reprocesses the vault search query and ensures all matching nodes have an encrypted secret,
// without modifying existing vault content or access rules.
//
// References:
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_refresh.rb
func (s *Service) Refresh(payload *Payload) (*RefreshResponse, error) {
	keyState, err := s.loadKeysCurrentState(payload)
	if err != nil {
		return nil, err
	}

	searchQuery := item_keys.NormalizeSearchQuery(keyState.SearchQuery)
	if searchQuery == nil {
		return nil, errors.New("search query is required")
	}

	refreshPayload := &Payload{
		VaultName:     payload.VaultName,
		VaultItemName: payload.VaultItemName,
		SearchQuery:   searchQuery,
	}

	searchedClients, err := s.getClientsFromSearch(refreshPayload)
	if err != nil {
		return nil, err
	}

	normalizedClients := item_keys.MergeClients(searchedClients, keyState.Clients)

	if payload.Clean {
		normalizedClients, _, err = cleanClients(normalizedClients, s.clientExists)
		if err != nil {
			return nil, err
		}
	}

	clientsEqual := item_keys.EqualLists(keyState.Clients, normalizedClients)

	refreshResponse := &RefreshResponse{
		Response: Response{
			URI: s.vaultURL(payload.VaultName),
		},
		KeysURIs: []string{},
	}

	// Skip re-encryption when requested and the effective client set is unchanged.
	if payload.SkipReencrypt && clientsEqual {
		return refreshResponse, nil
	}

	refreshContent, err := s.resolveUpdateContent(payload)
	if err != nil {
		return nil, err
	}

	refreshPayload.Admins = keyState.Admins
	refreshPayload.Clients = normalizedClients
	refreshPayload.Content = refreshContent
	refreshPayload.KeysMode = &keyState.Mode

	modeState := &item_keys.KeysModeState{
		Current: keyState.Mode,
		Desired: keyState.Mode,
	}
	keysResult, err := s.updateVault(refreshPayload, modeState)
	if err != nil {
		return nil, err
	}

	return &RefreshResponse{
		Response: Response{
			URI: s.vaultURL(payload.VaultName),
		},
		KeysURIs: keysResult.URIs,
	}, nil
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

// cleanClients partitions clients into those that still exist on the Chef server and those that do not.
func cleanClients(clients []string, exists func(string) (bool, error)) (kept []string, removed []string, err error) {
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

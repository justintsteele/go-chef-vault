package vault

import (
	"errors"
	"go-chef-vault/vault/item_keys"

	"github.com/go-chef/chef"
)

// RefreshResponse intentionally mirrors UpdateResponse for API parity
type RefreshResponse = UpdateResponse

// Refresh cleans and refreshes the vault, its clients, and admins
//
//	Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_refresh.rb
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
		SearchQuery: searchQuery,
	}

	searchedClients, err := s.getClientsFromSearch(refreshPayload)
	if err != nil {
		return nil, err
	}

	merged := item_keys.MergeClients(searchedClients, keyState.Clients)

	kept, _, err := cleanClients(merged, s.clientExists)
	if err != nil {
		return nil, err
	}

	normalizedClients := kept

	clientsEqual := item_keys.EqualLists(keyState.Clients, normalizedClients)

	refreshResponse := &RefreshResponse{
		Response: Response{
			URI: s.vaultURL(payload.VaultName),
		},
		KeysURIs: []string{},
	}

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

// clientExists performs a client lookup to validate the requested client still exists in the Chef Server
func (s *Service) clientExists(name string) (bool, error) {
	_, err := s.Client.Clients.Get(name)
	if err == nil {
		return true, nil
	}

	var chefErr *chef.ErrorResponse
	if errors.As(err, &chefErr) {
		if chefErr.Response.StatusCode == 404 {
			return false, nil
		}
	}

	return false, err
}

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

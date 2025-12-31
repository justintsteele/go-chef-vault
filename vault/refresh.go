package vault

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-chef/chef"
)

// RefreshResponse intentionally mirrors UpdateResponse for API parity
type RefreshResponse = UpdateResponse

// Refresh cleans and refreshes the vault, its clients, and admins
//
//	Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_refresh.rb
func (v *Service) Refresh(payload *VaultPayload) (*RefreshResponse, error) {
	keyState, err := v.loadKeysCurrentState(payload)
	if err != nil {
		return nil, err
	}

	searchQuery := normalizeSearchQuery(keyState.SearchQuery)
	if searchQuery == nil {
		return nil, errors.New("search query is required")
	}

	refreshPayload := &VaultPayload{
		SearchQuery: searchQuery,
	}

	searchedClients, err := v.getClientsFromSearch(refreshPayload)
	if err != nil {
		return nil, err
	}

	merged := mergeClients(searchedClients, keyState.Clients)

	kept, _, err := cleanClients(merged, v.clientExists)
	if err != nil {
		return nil, err
	}

	normalizedClients := kept

	clientsEqual := equalClientLists(keyState.Clients, normalizedClients)

	refreshResponse := &RefreshResponse{
		VaultResponse: VaultResponse{
			URI: v.vaultURL(payload.VaultName),
		},
		KeysURIs: []string{},
	}

	if payload.SkipReencrypt && clientsEqual {
		return refreshResponse, nil
	}

	refreshContent, err := resolveUpdateContent(v, payload)
	if err != nil {
		return nil, err
	}
	refreshPayload.Admins = keyState.Admins
	refreshPayload.Clients = normalizedClients
	refreshPayload.Content = refreshContent

	modeState := &KeysModeState{
		Current: keyState.Mode,
		Desired: keyState.Mode,
	}
	keysResult, err := v.updateVault(refreshPayload, modeState)
	if err != nil {
		return nil, err
	}

	return &RefreshResponse{
		VaultResponse: VaultResponse{
			URI: v.vaultURL(payload.VaultName),
		},
		KeysURIs: keysResult.URIs,
	}, nil
}

func equalClientLists(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	a = slices.Clone(a)
	b = slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

func normalizeSearchQuery(v any) *string {
	if v == nil {
		return nil
	}

	switch q := v.(type) {
	case *string:
		if q == nil || *q == "" {
			return nil
		}
		return q

	case string:
		if q == "" {
			return nil
		}
		return &q

	default:
		s := fmt.Sprint(v)
		if s == "" || s == "<nil>" {
			return nil
		}
		return &s
	}
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

func (v *Service) clientExists(name string) (bool, error) {
	_, err := v.Client.Clients.Get(name)
	if err == nil {
		return true, nil
	}

	if chefErr, ok := err.(*chef.ErrorResponse); ok {
		if chefErr.Response.StatusCode == 404 {
			return false, nil
		}
	}

	return false, err
}

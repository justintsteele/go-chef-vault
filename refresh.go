package vault

import (
	"errors"
	"maps"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

// RefreshResponse intentionally mirrors UpdateResponse for API parity.
type RefreshResponse = UpdateResponse

// refreshOps defines the callable operations required to execute a Refresh request.
type refreshOps struct {
	loadSharedSecret    func(*Payload) ([]byte, error)
	encryptSharedSecret func(pem string, secret []byte) (string, error)
	getItem             func(string, string) (chef.DataBagItem, error)
	updateVault         func(*Payload, *item_keys.KeysModeState) (*item_keys.VaultItemKeysResult, error)
}

// Refresh reprocesses the vault search query and ensures all matching nodes have an encrypted secret,
// without modifying existing vault content or access rules.
//
// References:
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_refresh.rb
func (s *Service) Refresh(payload *Payload) (*RefreshResponse, error) {
	ops := refreshOps{
		loadSharedSecret:    s.loadSharedSecret,
		encryptSharedSecret: item_keys.EncryptSharedSecret,
		getItem:             s.GetItem,
		updateVault:         s.updateVault,
	}

	return s.refresh(payload, ops)
}

// refresh is the worker called by the public API with the operational methods to complete the refresh request.
func (s *Service) refresh(payload *Payload, ops refreshOps) (*RefreshResponse, error) {
	keyState, err := s.loadKeysCurrentState(payload)
	if err != nil {
		return nil, err
	}

	nextState := &item_keys.VaultItemKeys{
		Id:          keyState.Id,
		Mode:        keyState.Mode,
		SearchQuery: keyState.SearchQuery,
		Admins:      append([]string(nil), keyState.Admins...),
		Clients:     append([]string(nil), keyState.Clients...),
		Keys:        maps.Clone(keyState.Keys),
	}

	searchQuery := item_keys.NormalizeSearchQuery(nextState.SearchQuery)
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

	normalizedClients := item_keys.MergeClients(searchedClients, nextState.Clients)

	addedClients := item_keys.DiffLists(normalizedClients, nextState.Clients)

	if payload.CleanUnknown {
		normalizedClients, _, err = s.cleanUnknownClients(payload, nextState, normalizedClients)
		if err != nil {
			return nil, err
		}
	}

	if payload.SkipReencrypt {
		return s.refreshSkipReencrypt(payload, nextState, addedClients, ops)
	}

	nextState.Clients = normalizedClients
	return s.refreshReencrypt(payload, nextState, ops)
}

func (s *Service) refreshReencrypt(payload *Payload, keyState *item_keys.VaultItemKeys, ops refreshOps) (*RefreshResponse, error) {
	currentItem, err := ops.getItem(payload.VaultName, payload.VaultItemName)
	if err != nil {
		return nil, err
	}

	currentDbi, err := item.DataBagItemMap(currentItem)
	if err != nil {
		return nil, err
	}

	payload.Content = currentDbi

	modeState := &item_keys.KeysModeState{
		Current: keyState.Mode,
		Desired: keyState.Mode,
	}

	keysResult, err := ops.updateVault(payload, modeState)
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

func (s *Service) refreshSkipReencrypt(payload *Payload, keyState *item_keys.VaultItemKeys, clients []string, ops refreshOps) (*RefreshResponse, error) {
	sharedSecret, err := ops.loadSharedSecret(payload)
	if err != nil {
		return nil, err
	}

	for _, actor := range clients {
		pub, err := s.clientPublicKey(actor)
		if err != nil {
			return nil, err
		}

		enc, err := ops.encryptSharedSecret(pub.PublicKey, sharedSecret)
		if err != nil {
			return nil, err
		}

		keyState.Keys[actor] = enc
	}

	keys := keyState.BuildKeysItem(clients)
	result := &item_keys.VaultItemKeysResult{}
	if err := s.writeKeys(payload, keyState.Mode, keys, result); err != nil {
		return nil, err
	}
	return &RefreshResponse{
		Response: Response{
			URI: s.vaultURL(payload.VaultName),
		},
		KeysURIs: result.URIs,
	}, nil
}

package vault

import (
	"maps"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

// RotateResponse represents the structure of the response from a RotateKeys operation.
type RotateResponse struct {
	Response
	KeysURIs []string `json:"keys_uris"`
}

// rotateOps defines the callable operations required to execute a RotateKeys request.
type rotateOps struct {
	getItem     func(string, string) (chef.DataBagItem, error)
	updateVault func(*Payload, *item_keys.KeysModeState) (*item_keys.VaultItemKeysResult, error)
}

// RotateKeys rotates the shared secret for a vault item by generating a new secret,
// re-encrypting all client/admin keys, and re-encrypting the vault data.
//
// References:
//   - Chef-vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_rotate_keys.rb
func (s *Service) RotateKeys(payload *Payload) (*RotateResponse, error) {
	if err := payload.validatePayload(); err != nil {
		return nil, err
	}

	ops := rotateOps{
		getItem:     s.GetItem,
		updateVault: s.updateVault,
	}
	return s.rotateKeys(payload, ops)
}

// rotateKeys is the worker called by the public API with the operational methods to complete a RotateKeys request.
func (s *Service) rotateKeys(payload *Payload, ops rotateOps) (*RotateResponse, error) {
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

	currentItem, err := ops.getItem(payload.VaultName, payload.VaultItemName)
	if err != nil {
		return nil, err
	}

	currentDbi, err := item.DataBagItemMap(currentItem)
	if err != nil {
		return nil, err
	}

	modeState := &item_keys.KeysModeState{
		Current: keyState.Mode,
		Desired: keyState.Mode,
	}

	query := item_keys.NormalizeSearchQuery(keyState.SearchQuery)

	rotatePayload := &Payload{
		VaultName:     payload.VaultName,
		VaultItemName: payload.VaultItemName,
		Content:       currentDbi,
		Admins:        keyState.Admins,
		SearchQuery:   query,
		KeysMode:      &keyState.Mode,
	}

	searchedClients, err := s.getClientsFromSearch(rotatePayload)
	if err != nil {
		return nil, err
	}

	normalizedClients := item_keys.MergeClients(searchedClients, nextState.Clients)

	if payload.CleanUnknown {
		normalizedClients, _, err = s.cleanUnknownClients(payload, nextState, normalizedClients)
		if err != nil {
			return nil, err
		}
	}

	rotatePayload.Clients = normalizedClients

	keysResult, err := ops.updateVault(rotatePayload, modeState)
	if err != nil {
		return nil, err
	}

	return &RotateResponse{
		Response: Response{
			URI: s.vaultURL(payload.VaultName),
		},
		KeysURIs: keysResult.URIs,
	}, nil
}

// RotateAllKeys performs a full key rotation for every vault item in the Chef server,
// regenerating shared secrets and re-encrypting data for each item.
//
// References:
//   - Chef-vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_rotate_all_keys.rb
func (s *Service) RotateAllKeys() ([]RotateResponse, error) {
	vaults, err := s.List()
	if err != nil {
		return nil, err
	}

	var res []RotateResponse
	for vault := range *vaults {
		vaultItems, err := s.ListItems(vault)
		if err != nil {
			return nil, err
		}
		for vaultItem := range *vaultItems {
			rotatePayload := &Payload{
				VaultName:     vault,
				VaultItemName: vaultItem,
			}

			result, err := s.RotateKeys(rotatePayload)
			if err != nil {
				return nil, err
			}
			res = append(res, *result)
		}
	}

	return res, nil
}

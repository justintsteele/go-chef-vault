package vault

// RotateKeys and RotateAllKeys are orchestration methods composed of
// lower-level primitives that are unit tested elsewhere. These functions
// intentionally do not have dedicated unit tests; behavior is validated
// via integration tests.

import (
	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

type RotateResponse struct {
	Response
	KeysURIs []string `json:"keys_uris"`
}

// RotateKeys rotates the shared secret for a vault item by generating a new secret,
// re-encrypting all client/admin keys, and re-encrypting the vault data.
//
// References:
//   - Chef-vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_rotate_keys.rb
func (s *Service) RotateKeys(payload *Payload) (*RotateResponse, error) {
	keyState, err := s.loadKeysCurrentState(payload)
	if err != nil {
		return nil, err
	}

	currentItem, err := s.GetItem(payload.VaultName, payload.VaultItemName)
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
		Clients:       keyState.Clients,
		Admins:        keyState.Admins,
		SearchQuery:   query,
		KeysMode:      &keyState.Mode,
	}

	keysResult, err := s.updateVault(rotatePayload, modeState)
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

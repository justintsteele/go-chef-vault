package vault

import (
	"fmt"
	"go-chef-vault/vault/item"
	"go-chef-vault/vault/item_keys"
)

// UpdateResponse represents the structure of the response from an Update operation.
type UpdateResponse struct {
	Response
	Data     *UpdateDataResponse `json:"data"`
	KeysURIs []string            `json:"keys_uris"`
}

// UpdateDataResponse represents the response returned after updating vault content.
type UpdateDataResponse struct {
	URI string `json:"uri"`
}

// Update modifies a vault item and its access keys on the Chef server.
//
// References:
//   - Chef API Docs: https://docs.chef.io/server/api_chef_server/#post-9
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_update.rb
func (s *Service) Update(payload *Payload) (*UpdateResponse, error) {
	keyState, err := s.loadKeysCurrentState(payload)
	if err != nil {
		return nil, err
	}

	payload.mergeKeyActors(keyState)

	finalQuery := item_keys.ResolveSearchQuery(keyState.SearchQuery, payload.SearchQuery)

	mode, modeState := payload.resolveKeysMode(keyState.Mode)
	keyState.Mode = mode

	content, err := s.resolveUpdateContent(payload)
	if err != nil {
		return nil, err
	}

	updatePayload := &Payload{
		VaultName:     payload.VaultName,
		VaultItemName: payload.VaultItemName,
		Content:       content,
		KeysMode:      &mode,
		SearchQuery:   finalQuery,
		Admins:        keyState.Admins,
		Clients:       keyState.Clients,
	}

	keysResult, err := s.updateVault(updatePayload, modeState)
	if err != nil {
		return nil, err
	}

	return &UpdateResponse{
		Response: Response{
			URI: s.vaultURL(updatePayload.VaultName),
		},
		Data: &UpdateDataResponse{
			URI: fmt.Sprintf(
				"%s/%s",
				s.vaultURL(updatePayload.VaultName),
				updatePayload.VaultItemName,
			),
		},
		KeysURIs: keysResult.URIs,
	}, nil
}

// updateVault performs the shared re-encryption logic used by Update and Refresh.
func (s *Service) updateVault(payload *Payload, modeState *item_keys.KeysModeState) (*item_keys.VaultItemKeysResult, error) {
	secret, err := item_keys.GenSecret(32)
	if err != nil {
		return nil, err
	}

	keysResult, err := s.createKeysDataBag(payload, modeState, secret, "update")
	if err != nil {
		return nil, err
	}

	encrypted, err := item.Encrypt(payload.VaultItemName, payload.Content, secret)
	if err != nil {
		return nil, err
	}

	if err := s.Client.DataBags.UpdateItem(
		payload.VaultName,
		payload.VaultItemName,
		&encrypted,
	); err != nil {
		return nil, err
	}

	return keysResult, nil
}

package vault

import (
	"fmt"

	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

// UpdateResponse represents the structure of the response from an Update operation.
type UpdateResponse struct {
	Response
	Data     *UpdateDataResponse `json:"data,omitempty"`
	KeysURIs []string            `json:"keys_uris,omitempty"`
}

// UpdateDataResponse represents the response returned after updating vault content.
type UpdateDataResponse struct {
	URI string `json:"uri"`
}

// updateOps defines the callable operations required to execute an Update request.
type updateOps struct {
	resolveUpdateContent func(p *Payload) (map[string]interface{}, error)
	updateVault          func(*Payload, *item_keys.KeysModeState) (*item_keys.VaultItemKeysResult, error)
}

// Update modifies a vault item and its access keys on the Chef server.
//
// References:
//   - Chef API Docs: https://docs.chef.io/server/api_chef_server/#post-9
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_update.rb
func (s *Service) Update(payload *Payload) (*UpdateResponse, error) {
	ops := updateOps{
		resolveUpdateContent: s.resolveUpdateContent,
		updateVault:          s.updateVault,
	}
	return s.update(payload, ops)
}

// update is the worker called by the public API with the operational methods to complete the update request.
func (s *Service) update(payload *Payload, ops updateOps) (*UpdateResponse, error) {
	keyState, err := s.loadKeysCurrentState(payload)
	if err != nil {
		return nil, err
	}

	keyState.Admins = item_keys.MergeClients(keyState.Admins, payload.Admins)
	keyState.Clients = item_keys.MergeClients(keyState.Clients, payload.Clients)

	if payload.Clean {
		if err := s.pruneKeys(keyState.Clients, keyState, payload); err != nil {
			return nil, err
		}
		keyState.Clients = nil
	}

	finalQuery := item_keys.ResolveSearchQuery(keyState.SearchQuery, payload.SearchQuery)

	mode, modeState := payload.resolveKeysMode(keyState.Mode)

	content, err := ops.resolveUpdateContent(payload)
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

	keysResult, err := ops.updateVault(updatePayload, modeState)
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

	keysResult, err := s.createKeysDataBag(payload, modeState, secret)
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

// resolveUpdateContent merges the payload content with the current content.
func (s *Service) resolveUpdateContent(p *Payload) (map[string]interface{}, error) {
	current, err := s.GetItem(p.VaultName, p.VaultItemName)
	if err != nil {
		return nil, err
	}

	currMap, err := item.DataBagItemMap(current)
	if err != nil {
		return nil, err
	}

	if p.Content == nil {
		return currMap, nil
	}

	merged, err := resolveContent(currMap, p.Content)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

// resolveContent merges the current contents to the vault with the contents provided by the payload.
func resolveContent(current, requested map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{}, len(current)+len(requested))

	for k, v := range current {
		out[k] = v
	}

	for k, v := range requested {
		out[k] = v
	}

	return out, nil
}

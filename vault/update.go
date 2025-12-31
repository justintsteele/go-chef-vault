package vault

import (
	"fmt"
	"go-chef-vault/vaultcrypto"
)

type UpdateResponse struct {
	VaultResponse
	Data     *UpdateDataResponse `json:"data"`
	KeysURIs []string            `json:"keys_uris"`
}

type KeysModeState struct {
	Current KeysMode `json:"current"`
	Desired KeysMode `json:"desired"`
}

type UpdateDataResponse struct {
	URI string `json:"uri"`
}

// Update modifies a vault on the server
//
//	Chef API Docs: https://docs.chef.io/server/api_chef_server/#post-9
//	Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_update.rb
func (v *Service) Update(payload *VaultPayload) (*UpdateResponse, error) {
	keyState, err := v.loadKeysCurrentState(payload)
	if err != nil {
		return nil, err
	}

	mergeKeyActors(keyState, payload)

	finalQuery := resolveSearchQuery(keyState.SearchQuery, payload.SearchQuery)

	mode, modeState := resolveKeysMode(keyState.Mode, payload)
	keyState.Mode = mode

	content, err := resolveUpdateContent(v, payload)
	if err != nil {
		return nil, err
	}

	updatePayload := &VaultPayload{
		VaultName:     payload.VaultName,
		VaultItemName: payload.VaultItemName,
		Content:       content,
		KeysMode:      &mode,
		SearchQuery:   finalQuery,
		Admins:        keyState.Admins,
		Clients:       keyState.Clients,
	}

	keysResult, err := v.updateVault(updatePayload, modeState)
	if err != nil {
		return nil, err
	}

	return &UpdateResponse{
		VaultResponse: VaultResponse{
			URI: v.vaultURL(updatePayload.VaultName),
		},
		Data: &UpdateDataResponse{
			URI: fmt.Sprintf(
				"%s/%s",
				v.vaultURL(updatePayload.VaultName),
				updatePayload.VaultItemName,
			),
		},
		KeysURIs: keysResult.URIs,
	}, nil
}

func (v *Service) updateVault(payload *VaultPayload, modeState *KeysModeState) (*VaultItemKeysResult, error) {
	secret, err := vaultcrypto.GenSecret(32)
	if err != nil {
		return nil, err
	}

	keysResult, err := v.createKeysDataBag(payload, modeState, secret, "update")
	if err != nil {
		return nil, err
	}

	encrypted, err := encryptContents(payload, secret)
	if err != nil {
		return nil, err
	}

	if err := v.Client.DataBags.UpdateItem(
		payload.VaultName,
		payload.VaultItemName,
		&encrypted,
	); err != nil {
		return nil, err
	}

	return keysResult, nil
}

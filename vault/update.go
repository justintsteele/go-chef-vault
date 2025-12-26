package vault

import (
	"encoding/json"
	"fmt"
	"go-chef-vault/crypto"
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
func (v *Service) Update(payload *VaultPayload) (result *UpdateResponse, err error) {
	keyState, err := v.loadKeysCurrentState(payload)
	if err != nil {
		return nil, err
	}

	if payload.Admins != nil {
		keyState.Admins = mergeClients(keyState.Admins, payload.Admins)
	}
	if payload.Clients != nil {
		if payload.Clean == true {
			keyState.Clients = payload.Clients
		} else {
			keyState.Clients = mergeClients(keyState.Clients, payload.Clients)
		}
	}

	var query *string
	switch q := effectiveSearchQuery(payload.SearchQuery).(type) {
	case string:
		query = &q
	default:
		query = nil
	}

	var keysModeState *KeysModeState
	if payload.KeysMode != nil {
		keysModeState = &KeysModeState{Current: keyState.Mode, Desired: *payload.KeysMode}
		keyState.Mode = *payload.KeysMode
	} else {
		keysModeState = &KeysModeState{Current: keyState.Mode, Desired: keyState.Mode}
	}
	mode := keyState.Mode

	requestedData, err := v.GetItem(payload.VaultName, payload.VaultItemName)
	if err != nil {
		return nil, err
	}

	if payload.Content != nil {
		mergedDataBagItem := make(map[string]any)
		for key, value := range payload.Content {
			plaintext, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			mergedDataBagItem[key] = plaintext
		}
		requestedData = mergedDataBagItem
	}

	requestedDataMap, err := dataBagItemMap(requestedData)
	if err != nil {
		return nil, err
	}

	updatePayload := &VaultPayload{
		VaultName:     payload.VaultName,
		VaultItemName: payload.VaultItemName,
		Content:       requestedDataMap,
		KeysMode:      &mode,
		SearchQuery:   query,
		Admins:        keyState.Admins,
		Clients:       keyState.Clients,
	}

	secret, err := crypto.GenSecret(32)
	if err != nil {
		return nil, err
	}

	result = &UpdateResponse{
		VaultResponse: VaultResponse{
			URI: v.vaultURL(payload.VaultName),
		},
	}

	keys, err := v.createKeysDataBag(updatePayload, keysModeState, secret, "update")

	if err != nil {
		return nil, err
	}

	result.KeysURIs = append(result.KeysURIs, keys.URIs...)

	eDB, eDBErr := encryptContents(payload, secret)
	if eDBErr != nil {
		return nil, eDBErr
	}

	if eDbErr := v.Client.DataBags.UpdateItem(payload.VaultName, payload.VaultItemName, &eDB); eDbErr != nil {
		return nil, eDbErr
	}

	result.Data = &UpdateDataResponse{URI: fmt.Sprintf("%s/%s", v.vaultURL(payload.VaultName), payload.VaultItemName)}

	return result, nil
}

// loadKeysCurrentState loads the current state of the base keys data bag item
func (v *Service) loadKeysCurrentState(payload *VaultPayload) (*VaultItemKeys, error) {
	baseKeys, err := v.Client.DataBags.GetItem(payload.VaultName, payload.VaultItemName+"_keys")
	if err != nil {
		return nil, err
	}

	baseKeysData, err := json.Marshal(baseKeys)
	if err != nil {
		return nil, err
	}
	var vik VaultItemKeys
	if err := json.Unmarshal(baseKeysData, &vik); err != nil {
		return nil, err
	}

	return &vik, nil
}

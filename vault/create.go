package vault

import (
	"fmt"
	"go-chef-vault/vault/item"
	"go-chef-vault/vault/item_keys"

	"github.com/go-chef/chef"
)

type CreateResponse struct {
	Response
	Data     *CreateDataResponse `json:"data"`
	KeysURIs []string            `json:"keys_uris"`
}

type CreateDataResponse struct {
	URI string `json:"uri"`
}

// Create adds a vault item and its associated keys to the Chef server.
//
// References:
//   - Chef API Docs: https://docs.chef.io/server/api_chef_server/#post-9
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_create.rb
func (s *Service) Create(payload *Payload) (result *CreateResponse, err error) {
	vaultDataBag := chef.DataBag{
		Name: payload.VaultName,
	}

	_, vDbErr := s.Client.DataBags.Create(&vaultDataBag)
	if vDbErr != nil {
		return nil, vDbErr
	}

	result = &CreateResponse{
		Response: Response{
			URI: s.vaultURL(payload.VaultName),
		},
	}

	secret, err := item_keys.GenSecret(32)
	if err != nil {
		return nil, err
	}

	keysModeState := &item_keys.KeysModeState{
		Current: payload.effectiveKeysMode(),
		Desired: payload.effectiveKeysMode(),
	}

	keys, err := s.createKeysDataBag(payload, keysModeState, secret, "create")

	if err != nil {
		return nil, err
	}

	result.KeysURIs = append(result.KeysURIs, keys.URIs...)

	eDB, eDBErr := item.Encrypt(payload.VaultItemName, payload.Content, secret)
	if eDBErr != nil {
		return nil, eDBErr
	}

	if eDbErr := s.Client.DataBags.CreateItem(payload.VaultName, &eDB); eDbErr != nil {
		return nil, eDbErr
	}

	result.Data = &CreateDataResponse{URI: fmt.Sprintf("%s/%s", s.vaultURL(payload.VaultName), payload.VaultItemName)}

	return result, nil
}

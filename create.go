package vault

import (
	"fmt"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

// CreateResponse represents the structure of the response from a Create operation.
type CreateResponse struct {
	Response
	Data     *CreateDataResponse `json:"data"`
	KeysURIs []string            `json:"keys_uris"`
}

// CreateDataResponse represents the response returned after creating vault content.
type CreateDataResponse struct {
	URI string `json:"uri"`
}

// createOps defines the callable operations required to execute an Create request.
type createOps struct {
	createKeysDataBag func(*Payload, *item_keys.KeysModeState, []byte) (*item_keys.VaultItemKeysResult, error)
}

// Create adds a vault item and its associated keys to the Chef server.
//
// References:
//   - Chef API Docs: https://docs.chef.io/server/api_chef_server/#post-9
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_create.rb
func (s *Service) Create(payload *Payload) (*CreateResponse, error) {
	if err := payload.validatePayload(); err != nil {
		return nil, err
	}

	ops := createOps{
		createKeysDataBag: s.createKeysDataBag,
	}

	return s.create(payload, ops)
}

// create is the worker called by the public API with the operational methods to complete the create request.
func (s *Service) create(payload *Payload, ops createOps) (*CreateResponse, error) {
	vaultDataBag := chef.DataBag{
		Name: payload.VaultName,
	}

	_, err := s.Client.DataBags.Create(&vaultDataBag)
	if err != nil {
		return nil, err
	}

	result := &CreateResponse{
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

	keys, err := ops.createKeysDataBag(payload, keysModeState, secret)
	if err != nil {
		return nil, err
	}

	result.KeysURIs = append(result.KeysURIs, keys.URIs...)

	eDB, err := item.Encrypt(payload.VaultItemName, payload.Content, secret)
	if err != nil {
		return nil, err
	}

	if err := s.Client.DataBags.CreateItem(payload.VaultName, &eDB); err != nil {
		return nil, err
	}

	result.Data = &CreateDataResponse{URI: fmt.Sprintf("%s/%s", s.vaultURL(payload.VaultName), payload.VaultItemName)}

	return result, nil
}

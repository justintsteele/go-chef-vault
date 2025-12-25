package vault

import (
	"fmt"
	"go-chef-vault/crypto"

	"github.com/go-chef/chef"
)

type CreateResponse struct {
	VaultResponse
	Data     *CreateDataResponse `json:"data"`
	KeysURIs []string            `json:"keys_uris"`
}

type CreateDataResponse struct {
	URI string `json:"uri"`
}

// Create adds a vault to the server
//
//	Chef API Docs: https://docs.chef.io/server/api_chef_server/#post-9
//	Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_create.rb
func (v *Service) Create(payload *VaultPayload) (result *CreateResponse, err error) {
	vaultDataBag := chef.DataBag{
		Name: payload.VaultName,
	}

	_, vDbErr := v.Client.DataBags.Create(&vaultDataBag)
	if vDbErr != nil {
		return nil, vDbErr
	}

	result = &CreateResponse{
		VaultResponse: VaultResponse{
			URI: v.vaultURL(payload.VaultName),
		},
	}

	secret, err := crypto.GenSecret(32)
	if err != nil {
		return nil, err
	}

	keysModeState := &KeysModeState{
		Current: payload.EffectiveKeysMode(),
		Desired: payload.EffectiveKeysMode(),
	}

	keys, err := v.createKeysDataBag(payload, keysModeState, secret, "create")

	if err != nil {
		return nil, err
	}

	result.KeysURIs = append(result.KeysURIs, keys.URIs...)

	eDB, eDBErr := encryptContents(payload, secret)
	if eDBErr != nil {
		return nil, eDBErr
	}

	if eDbErr := v.Client.DataBags.CreateItem(payload.VaultName, &eDB); eDbErr != nil {
		return nil, eDbErr
	}

	result.Data = &CreateDataResponse{URI: fmt.Sprintf("%s/%s", v.vaultURL(payload.VaultName), payload.VaultItemName)}

	return result, nil
}

package vault

import (
	"fmt"
)

type DeleteResponse struct {
	VaultResponse
	KeysURIs []string `json:"keys,omitempty"`
}

// Delete removes the entire vault, all the items, and keys from the server (nuclear option)
//
//	Chef API Docs: https://docs.chef.io/api_chef_server/#delete-9
func (v *Service) Delete(name string) (result *DeleteResponse, err error) {
	vaultUri := fmt.Sprintf("%s", v.vaultURL(name))
	_, err = v.Client.DataBags.Delete(name)
	result = &DeleteResponse{
		VaultResponse: VaultResponse{
			vaultUri,
		},
	}
	return
}

// DeleteItem removes a specified item from a vault and its keys
//
//	Chef API Docs: https://docs.chef.io/api_chef_server/#delete-10
//	Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_delete.rb
func (v *Service) DeleteItem(name string, item string) (resp *DeleteResponse, err error) {
	itemUri := fmt.Sprintf("%s/%s", v.vaultURL(name), item)
	resp = &DeleteResponse{
		VaultResponse: VaultResponse{
			URI: itemUri,
		},
	}
	vErr := v.Client.DataBags.DeleteItem(name, item)

	// TODO: this needs to account for keys mode == sparse and delete the sparse keys as well
	itemKeysUri := fmt.Sprintf("%s/%s", v.vaultURL(name), item+"_keys")
	kErr := v.Client.DataBags.DeleteItem(name, item+"_keys")
	if vErr != nil || kErr != nil {
		err = fmt.Errorf("vault error: %v, keys error: %v", vErr, kErr)
	}
	resp.KeysURIs = append(resp.KeysURIs, itemKeysUri)
	return
}

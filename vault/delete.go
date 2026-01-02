package vault

import (
	"fmt"
)

type DeleteResponse struct {
	Response
	KeysURIs []string `json:"keys,omitempty"`
}

// Delete removes the entire vault, all the items, and keys from the server (nuclear option)
//
//	Chef API Docs: https://docs.chef.io/api_chef_server/#delete-9
func (s *Service) Delete(name string) (result *DeleteResponse, err error) {
	vaultUri := fmt.Sprintf("%s", s.vaultURL(name))
	_, err = s.Client.DataBags.Delete(name)
	result = &DeleteResponse{
		Response: Response{
			vaultUri,
		},
	}
	return
}

// DeleteItem removes a specified item from a vault and its keys
//
//	Chef API Docs: https://docs.chef.io/api_chef_server/#delete-10
//	Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_delete.rb
func (s *Service) DeleteItem(name string, item string) (resp *DeleteResponse, err error) {
	itemUri := fmt.Sprintf("%s/%s", s.vaultURL(name), item)
	resp = &DeleteResponse{
		Response: Response{
			URI: itemUri,
		},
	}
	vErr := s.Client.DataBags.DeleteItem(name, item)

	// TODO: this needs to account for keys mode == sparse and delete the sparse keys as well
	itemKeysUri := fmt.Sprintf("%s/%s", s.vaultURL(name), item+"_keys")
	kErr := s.Client.DataBags.DeleteItem(name, item+"_keys")
	if vErr != nil || kErr != nil {
		err = fmt.Errorf("vault error: %s, keys error: %s", vErr, kErr)
	}
	resp.KeysURIs = append(resp.KeysURIs, itemKeysUri)
	return
}

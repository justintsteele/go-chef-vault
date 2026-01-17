package vault

import (
	"fmt"

	"github.com/justintsteele/go-chef-vault/item_keys"
)

// DeleteResponse represents the structure of the response from a Delete operation.
type DeleteResponse struct {
	Response
	KeysURIs []string `json:"keys,omitempty"`
}

// Delete destroys the entire vault, all the items, and keys from the Chef Server (nuclear option).
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#delete-9
func (s *Service) Delete(name string) (result *DeleteResponse, err error) {
	vaultUri := s.vaultURL(name)
	_, err = s.Client.DataBags.Delete(name)
	result = &DeleteResponse{
		Response: Response{
			vaultUri,
		},
	}
	return
}

// DeleteItem destroys a specified vault item and its keys.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#delete-10
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_delete.rb
func (s *Service) DeleteItem(name string, item string) (resp *DeleteResponse, err error) {
	payload := &Payload{
		VaultName:     name,
		VaultItemName: item,
	}
	keyState, err := s.loadKeysCurrentState(payload)
	if err != nil {
		return
	}

	// the vault item is deleted first, followed by best-effort key cleanup.
	resp, err = s.deleteVaultItem(name, item)
	if err != nil {
		return
	}

	if keyState.Mode == item_keys.KeysModeSparse {
		actors := make([]string, len(keyState.Admins)+len(keyState.Clients))
		actors = append(actors, keyState.Admins...)
		actors = append(actors, keyState.Clients...)
		if err := s.deleteSparseKeys(name, item, actors, resp); err != nil {
			return nil, err
		}
	}

	if err := s.deleteDefaultKeys(name, item, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// deleteVaultItem removes the encrypted data bag portion of the vault.
func (s *Service) deleteVaultItem(name string, item string) (resp *DeleteResponse, err error) {
	itemUri := fmt.Sprintf("%s/%s", s.vaultURL(name), item)
	if err := s.Client.DataBags.DeleteItem(name, item); err != nil {
		return nil, err
	}
	resp = &DeleteResponse{
		Response: Response{
			URI: itemUri,
		},
	}
	return
}

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
func (s *Service) Delete(vaultName string) (*DeleteResponse, error) {
	if vaultName == "" {
		return nil, ErrMissingVaultName
	}

	vaultUri := s.vaultURL(vaultName)
	_, err := s.Client.DataBags.Delete(vaultName)
	if err != nil {
		return nil, err
	}
	return &DeleteResponse{
		Response: Response{
			vaultUri,
		},
	}, nil
}

// DeleteItem destroys a specified vault item and its keys.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#delete-10
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_delete.rb
func (s *Service) DeleteItem(vaultName, vaultItem string) (*DeleteResponse, error) {
	pl := &Payload{
		VaultName:     vaultName,
		VaultItemName: vaultItem,
	}

	if err := pl.validatePayload(); err != nil {
		return nil, err
	}

	keyState, err := s.loadKeysCurrentState(pl)
	if err != nil {
		return nil, err
	}

	// the vault item is deleted first, followed by best-effort key cleanup.
	resp, err := s.deleteVaultItem(pl.VaultName, pl.VaultItemName)
	if err != nil {
		return nil, err
	}

	if keyState.Mode == item_keys.KeysModeSparse {
		actors := make([]string, len(keyState.Admins)+len(keyState.Clients))
		actors = append(actors, keyState.Admins...)
		actors = append(actors, keyState.Clients...)
		if err := s.deleteSparseKeys(pl.VaultName, pl.VaultItemName, actors, resp); err != nil {
			return nil, err
		}
	}

	if err := s.deleteDefaultKeys(pl.VaultName, pl.VaultItemName, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// deleteVaultItem removes the encrypted data bag portion of the vault.
func (s *Service) deleteVaultItem(vaultName, vaultItem string) (*DeleteResponse, error) {
	itemUri := fmt.Sprintf("%s/%s", s.vaultURL(vaultName), vaultItem)
	if err := s.Client.DataBags.DeleteItem(vaultName, vaultItem); err != nil {
		return nil, err
	}
	return &DeleteResponse{
		Response: Response{
			URI: itemUri,
		},
	}, nil
}

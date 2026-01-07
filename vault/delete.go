package vault

import (
	"fmt"

	"github.com/justintsteele/go-chef-vault/vault/cheferr"
	"github.com/justintsteele/go-chef-vault/vault/item_keys"
)

type DeleteResponse struct {
	Response
	KeysURIs []string `json:"keys,omitempty"`
}

// Delete removes the entire vault, all the items, and keys from the Chef Server (nuclear option).
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

// DeleteItem removes a specified item from a vault and its keys.
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

	switch keyState.Mode {
	case item_keys.KeysModeDefault:
		if err := s.deleteDefaultKeys(name, item, resp); err != nil {
			return nil, err
		}
	case item_keys.KeysModeSparse:
		if err := s.deleteSparseKeys(name, item, keyState, resp); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// deleteVaultItem removes the encrypted databag portion of the vault.
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

// deleteDefaultKeys removes the base keys and any actor keys stored in default mode.
func (s *Service) deleteDefaultKeys(name string, item string, out *DeleteResponse) (err error) {
	itemKeysUri := fmt.Sprintf("%s/%s", s.vaultURL(name), item+"_keys")
	if err := s.Client.DataBags.DeleteItem(name, item+"_keys"); err != nil {
		return err
	}
	out.KeysURIs = append(out.KeysURIs, itemKeysUri)
	return
}

// deleteSparseKeys removes all actor keys and the base sparse keys item.
func (s *Service) deleteSparseKeys(name string, item string, keyState *item_keys.VaultItemKeys, out *DeleteResponse) (err error) {
	baseKeyId := fmt.Sprintf("%s_keys", item)
	baseUri := fmt.Sprintf("%s/%s", s.vaultURL(name), baseKeyId)
	if err := s.Client.DataBags.DeleteItem(name, baseKeyId); err != nil {
		if !cheferr.IsNotFound(err) {
			return err
		}
	}
	out.KeysURIs = append(out.KeysURIs, baseUri)

	actors := item_keys.MergeClients(keyState.Admins, keyState.Clients)
	for _, actor := range actors {
		sparseId := fmt.Sprintf("%s_key_%s", item, actor)
		adminKeyUri := fmt.Sprintf("%s/%s", s.vaultURL(name), sparseId)
		if err := s.Client.DataBags.DeleteItem(name, sparseId); err != nil {
			if !cheferr.IsNotFound(err) {
				return err
			}
		}
		out.KeysURIs = append(out.KeysURIs, adminKeyUri)
	}
	return
}

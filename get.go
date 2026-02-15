package vault

import (
	"crypto/rsa"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

// getOps defines the callable operations required to execute a GetItem request.
type getOps struct {
	deriveAESKey func(string, *rsa.PrivateKey) ([]byte, error)
	decrypt      func(chef.DataBagItem, []byte) (chef.DataBagItem, error)
}

// GetItem returns the decrypted items in the vault.
//
// References:
//   - Chef API Docs: https://docs.chef.io/api_chef_server/#get-26
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_show.rb
func (s *Service) GetItem(vaultName, vaultItem string) (chef.DataBagItem, error) {
	pl := &Payload{
		VaultName:     vaultName,
		VaultItemName: vaultItem,
	}

	if err := pl.validatePayload(); err != nil {
		return nil, err
	}

	ops := getOps{
		deriveAESKey: item_keys.DeriveAESKey,
		decrypt:      item.Decrypt,
	}
	return s.getItem(pl.VaultName, pl.VaultItemName, ops)
}

// getItem is the worker called by the public API with the operational methods to complete the update request.
func (s *Service) getItem(vaultName, vaultItem string, ops getOps) (chef.DataBagItem, error) {
	actorKey, err := s.loadActorKey(vaultName, vaultItem)
	if err != nil {
		return nil, err
	}

	aesKey, err := ops.deriveAESKey(
		actorKey,
		s.Client.Auth.PrivateKey,
	)
	if err != nil {
		return nil, err
	}

	rawItem, err := s.Client.DataBags.GetItem(vaultName, vaultItem)
	if err != nil {
		return nil, err
	}

	return ops.decrypt(rawItem, aesKey)
}

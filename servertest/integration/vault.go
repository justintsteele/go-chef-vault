package integration

// Test the go-chef/chef/ chef server api /chef-vault (databags) endpoints against a live server or goiardi

import (
	"github.com/justintsteele/go-chef-vault/vault"
)

func RunVault(cfg Config) error {
	client := mustCreateClient(cfg)
	service := vault.NewService(client)

	runStep("Create Vault", func() (any, error) {
		return createVault(service)
	})

	runStep("Get Item", func() (any, error) {
		return getVault(service)
	})

	runStep("Update Item", func() (any, error) {
		return updateContent(service)
	})

	runStep("Get Item", func() (any, error) {
		return getVault(service)
	})

	runStep("List Items", func() (any, error) {
		return listItems(service)
	})

	runStep("List Vaults", func() (any, error) {
		return list(service)
	})

	runStep("Delete Items", func() (any, error) {
		return deleteItem(service)
	})

	runStep("Delete Vault", func() (any, error) {
		return deleteVault(service)
	})

	return nil
}

func runStep[T any](name string, fn func() (T, error)) {
	res, err := fn()
	Must(err)
	report(name, res)
}

package integration

// Test the go-chef/chef/ chef server api /chef-vault (databags) endpoints against a live server or goiardi

import (
	"fmt"

	"github.com/justintsteele/go-chef-vault/vault"
)

func RunVault(cfg Config) error {
	if cfg.Target == TargetGoiardi {
		cfg.Knife = fmt.Sprintf("%s/%s.rb", cfg.WorkDir, goiardiUser)
		defer func() {
			if !cfg.Keep {
				cfg.Knife = fmt.Sprintf("%s/%s.rb", cfg.WorkDir, goiardiAdminUser)
				client := cfg.mustCreateClient()
				service := vault.NewService(client)

				runStep("Delete Items", func() (any, error) {
					return deleteItem(service)
				})

				runStep("Delete Vault", func() (any, error) {
					return deleteVault(service)
				})

				runStep("Delete User", func() (any, error) {
					return deleteUser(service)
				})
			}
		}()
	}

	client := cfg.mustCreateClient()
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

	runStep("Get Item Keys", func() (any, error) {
		return getKeys(service)
	})

	runStep("List Items", func() (any, error) {
		return listItems(service)
	})

	runStep("List Vaults", func() (any, error) {
		return list(service)
	})

	runStep("Refresh Vaults", func() (any, error) {
		return refresh(service)
	})

	return nil
}

func runStep[T any](name string, fn func() (T, error)) {
	res, err := fn()
	Must(err)
	report(name, res)
}

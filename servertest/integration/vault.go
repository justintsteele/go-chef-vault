package integration

// Test the go-chef/chef/ chef server api /chef-vault (databags) endpoints against a live server or goiardi

import (
	"fmt"

	"github.com/justintsteele/go-chef-vault/vault"
)

type IntegrationService struct {
	Service *vault.Service
}

func NewIntegrationService(service *vault.Service) *IntegrationService {
	return &IntegrationService{
		Service: service,
	}
}

func RunVault(cfg Config) error {
	if cfg.Target == TargetGoiardi {
		cfg.Knife = fmt.Sprintf("%s/%s.rb", cfg.WorkDir, goiardiUser)
		defer func() {
			if !cfg.Keep {
				cfg.Knife = fmt.Sprintf("%s/%s.rb", cfg.WorkDir, goiardiAdminUser)
				client := cfg.mustCreateClient()
				service := vault.NewService(client)
				isvc := NewIntegrationService(service)

				runStep("Delete Items", func() (any, error) {
					return isvc.deleteItem()
				})

				runStep("Delete Vault", func() (any, error) {
					return isvc.deleteVault()
				})

				runStep("Delete User", func() (any, error) {
					return isvc.deleteUser()
				})
			}
		}()
	}

	client := cfg.mustCreateClient()
	service := vault.NewService(client)
	isvc := NewIntegrationService(service)

	runStep("Create Vault", func() (any, error) {
		return isvc.createVault()
	})

	runStep("Get Item", func() (any, error) {
		return isvc.getVault()
	})

	runStep("Update Item", func() (any, error) {
		return isvc.updateContent()
	})

	runStep("Get Item", func() (any, error) {
		return isvc.getVault()
	})

	runStep("Get Item Keys", func() (any, error) {
		return isvc.getKeys()
	})

	runStep("List Items", func() (any, error) {
		return isvc.listItems()
	})

	runStep("List Vaults", func() (any, error) {
		return isvc.list()
	})

	runStep("Refresh Vaults", func() (any, error) {
		return isvc.refresh()
	})

	return nil
}

func runStep[T any](name string, fn func() (T, error)) {
	res, err := fn()
	Must(err)
	report(name, res)
}

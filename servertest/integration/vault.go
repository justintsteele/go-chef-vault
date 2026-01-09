package integration

// Test the go-chef/chef/ chef server api /chef-vault (databags) endpoints against a live server or goiardi

import (
	"fmt"

	"github.com/justintsteele/go-chef-vault/vault"
)

type IntegrationService struct {
	Service *vault.Service
}

const (
	newNodeName   = "testhost1"
	fakeNodeName  = "fakehost1"
	vaultName     = "go-vault1"
	vaultItemName = "secret1"
)

func NewIntegrationService(service *vault.Service) *IntegrationService {
	return &IntegrationService{
		Service: service,
	}
}

func RunVault(cfg Config) error {
	if cfg.Target == TargetGoiardi {
		cfg.Knife = fmt.Sprintf("%s/%s.rb", cfg.WorkDir, goiardiUser)
		defer func() {
			client := cfg.mustCreateClient()
			service := vault.NewService(client)
			isvc := NewIntegrationService(service)

			if !cfg.Keep {
				cfg.Knife = fmt.Sprintf("%s/%s.rb", cfg.WorkDir, goiardiAdminUser)

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

	defer func() {
		client := cfg.mustCreateClient()
		service := vault.NewService(client)
		isvc := NewIntegrationService(service)

		// Delete our data bag fixtures
		Must(isvc.deleteDataBags())

		// Delete our new node and client so we can re-run the tests without panics
		Must(isvc.deleteClients(newNodeName))
		Must(isvc.deleteClients(fakeNodeName))
	}()

	client := cfg.mustCreateClient()
	service := vault.NewService(client)
	isvc := NewIntegrationService(service)

	runStep("Create Vault", func() (any, error) {
		return isvc.createVault()
	})

	runStep("Is Vault?", func() (any, error) {
		return isvc.isVault()
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

	runStep("Remove", func() (any, error) {
		return isvc.remove()
	})
	return nil
}

func runStep[T any](name string, fn func() (T, error)) {
	res, err := fn()
	Must(err)
	report(name, res)
}

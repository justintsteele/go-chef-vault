package integration

// Test the go-chef/chef/ chef server api /chef-vault (databags) endpoints against a live server or goiardi

import (
	"fmt"

	"github.com/justintsteele/go-chef-vault"
)

type IntegrationService struct {
	Service *vault.Service
}

const (
	newNodeName    = "testhost1"
	fakeNodeName   = "fakehost1"
	vaultName      = "go-vault1"
	vaultItemName  = "secret1"
	vault2Name     = "go-vault2"
	vault2ItemName = "secret2"
)

func NewIntegrationService(service *vault.Service) *IntegrationService {
	return &IntegrationService{
		Service: service,
	}
}

func RunScenarios(cfg Config, reporter ScenarioReporter) (results []*ExecutedScenario, err error) {
	defer func() {
		if reporter != nil {
			reporter.Report(results)
		}
	}()

	if cfg.Target == TargetGoiardi {
		cfg.Knife = fmt.Sprintf("%s/%s.rb", cfg.WorkDir, goiardiUser)
		defer func() {
			client := cfg.mustCreateClient()
			service := vault.NewService(client)
			isvc := NewIntegrationService(service)

			if !cfg.Keep {
				cfg.Knife = fmt.Sprintf("%s/%s.rb", cfg.WorkDir, goiardiAdminUser)

				deferScenarios := []Scenario{
					deleteItem(),
					deleteVault(vaultName, vaultItemName),
					deleteVault(vault2Name, vault2ItemName),
				}

				for _, s := range deferScenarios {
					result := s.Run(isvc)
					results = append(results, &ExecutedScenario{
						Name:   s.Name,
						Result: result,
					})
				}

				Must(isvc.deleteUser())
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
		Must(isvc.deleteClients(newNodeName + "0"))
	}()

	client := cfg.mustCreateClient()
	service := vault.NewService(client)
	isvc := NewIntegrationService(service)

	scenarios := []Scenario{
		createVault(),
		isItem(),
		getVault(),
		getKeys(),
		updateSparse(),
		list(),
		refreshSkipReencrypt(),
		refreshReencrypt(),
		rotate(),
		rotateAll(),
		remove(),
		updateDefault(),
	}

	for _, s := range scenarios {
		result := s.Run(isvc)
		results = append(results, &ExecutedScenario{
			Name:   s.Name,
			Result: result,
		})
	}

	return results, nil
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

package integration

import (
	"fmt"

	vault "github.com/justintsteele/go-chef-vault"
	"github.com/justintsteele/go-chef-vault/item"
)

func getKeys() Scenario {
	return Scenario{
		Name: "Get Keys",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}

			result, err := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_keys")
			sr.assertNoError("Get Default Keys", err)

			dbi, dbierr := item.DataBagItemMap(result)
			sr.assertNoError("Map default keys", dbierr)

			assertDefaultKeysOnly(sr, i.Service, vaultName, vaultItemName, dbi, "admins")

			return sr
		},
	}
}

func assertSparseKeysOnly(sr *ScenarioResult, svc *vault.Service, vaultName, vaultItemName string, keysDbi map[string]any, actorsKey string) {
	assertKeyMode(sr, svc, vaultName, vaultItemName, keysDbi, actorsKey, SparseOnly)
}

func assertDefaultKeysOnly(sr *ScenarioResult, svc *vault.Service, vaultName, vaultItemName string, keysDbi map[string]any, actorsKey string) {
	assertKeyMode(sr, svc, vaultName, vaultItemName, keysDbi, actorsKey, DefaultOnly)
}

func assertKeyMode(sr *ScenarioResult, svc *vault.Service, vaultName, vaultItemName string, keysDbi map[string]any, actorsKey string, mode KeyMode) {
	rawActors, ok := keysDbi[actorsKey].([]interface{})
	if !ok {
		sr.assert(fmt.Sprintf("%s list present", actorsKey), false, fmt.Errorf("expected %s to be []interface{}, got %T", actorsKey, keysDbi[actorsKey]))
		return
	}

	for _, a := range rawActors {
		actor, ok := a.(string)
		if !ok {
			sr.assert(fmt.Sprintf("actor is string in %s", actorsKey), false, fmt.Errorf("expected string actor, got %T", a))
			continue
		}

		sparseID := vaultItemName + "_key_" + actor

		_, sparseErr := svc.Client.DataBags.GetItem(vaultName, sparseID)
		_, defaultExists := keysDbi[actor]

		switch mode {
		case SparseOnly:
			sr.assertNoError(fmt.Sprintf("sparse key exists for %s", actor), sparseErr)
			sr.assert(fmt.Sprintf("default key does not exist for %s", actor), !defaultExists, fmt.Errorf("unexpected default key present for %s", actor))

		case DefaultOnly:
			sr.assert(fmt.Sprintf("sparse key does not exist for %s", actor), sparseErr != nil, fmt.Errorf("unexpected sparse key present for %s", actor))
			sr.assert(fmt.Sprintf("default key exists for %s", actor), defaultExists, fmt.Errorf("default key for %s missing", actor))
		}
	}
}

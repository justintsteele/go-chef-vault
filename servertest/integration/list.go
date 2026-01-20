package integration

func list() Scenario {
	return Scenario{
		Name: "List",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}

			result, err := i.Service.List()
			sr.assertNoError("list vaults", err)
			sr.assertEqual("number of vaults", len(*result), 2)

			res, err := i.Service.ListItems(vaultName)
			sr.assertNoError("list vault items", err)
			sr.assertEqual("number of vaults items", len(*res), 1)

			return sr
		},
	}
}

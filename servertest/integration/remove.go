package integration

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/justintsteele/go-chef-vault"
	"github.com/justintsteele/go-chef-vault/item"
)

func remove() Scenario {
	return Scenario{
		Name: "Remove",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}

			// fixtures: if these fail, scenario can't proceed meaningfully
			var raw map[string]interface{}
			vaultItem := `{"fuz": { "buz": "buz-value-5" } }`
			if err := json.Unmarshal([]byte(vaultItem), &raw); err != nil {
				sr.assertNoError("unmarshal remove content", err)
				return sr
			}

			pl := &vault.Payload{
				VaultName:     vaultName,
				VaultItemName: vaultItemName,
				Clients:       []string{fakeNodeName},
				CleanUnknown:  true,
				Content:       raw,
			}

			preKeys, _ := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_keys")
			preKeysDbi, _ := item.DataBagItemMap(preKeys)
			sr.assertContains(fmt.Sprintf("pre-remove clients include %s", fakeNodeName), preKeysDbi["clients"], fakeNodeName)

			preData, _ := i.Service.GetItem(vaultName, vaultItemName)
			preDataDbi, _ := item.DataBagItemMap(preData)
			preCheck := preDataDbi["fuz"].(map[string]interface{})["buz"]
			sr.assert("pre-check data found", preCheck != nil, errors.New("pre-check data not found"))

			_, err := i.Service.Remove(pl)
			sr.assertNoError("remove vault", err)

			postKeys, _ := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_keys")
			postKeysDbi, _ := item.DataBagItemMap(postKeys)
			sr.assertNotContains(fmt.Sprintf("post-remove clients should not include %s", fakeNodeName), postKeysDbi["clients"], fakeNodeName)

			var rem map[string]interface{}
			remItem := `{
"id": "secret1",
"baz": "baz-value-1", 
"foo": "foo-value-3",
"fuz": {
	  "faz": "faz-value-4",
	  "boz": "boz-value-6"
	}
}`
			if err := json.Unmarshal([]byte(remItem), &rem); err != nil {
				sr.assertNoError("unmarshal remove content", err)
				return sr
			}

			postData, _ := i.Service.GetItem(vaultName, vaultItemName)
			postDataDbi, _ := item.DataBagItemMap(postData)
			postCheck := postDataDbi["fuz"].(map[string]interface{})["buz"]
			sr.assert("post-check data not found", postCheck == nil, errors.New("post-check data found"))

			sr.assertEqual("full data bag item post remove", postDataDbi, rem)

			_, err = i.Service.Remove(nil)
			sr.assertError(fmt.Sprintf("nil payload: %v", err), err)

			return sr
		},
	}
}

package integration

import (
	"fmt"

	"github.com/justintsteele/go-chef-vault"
	"github.com/justintsteele/go-chef-vault/item"
)

func refresh() Scenario {
	return Scenario{
		Name: "Refresh",
		Run: func(i *IntegrationService) *ScenarioResult {
			sr := &ScenarioResult{}
			pl := &vault.Payload{
				VaultName:     vaultName,
				VaultItemName: vaultItemName,
				CleanUnknown:  true,
				SkipReencrypt: true,
			}

			// gather preflight information to check after the refresh.
			preKeys, _ := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_keys")
			preKeysDbi, _ := item.DataBagItemMap(preKeys)
			sr.assertEqual("pre-refresh clients", preKeysDbi["clients"], []string{"fakehost1", "testhost1", "testhost11"})

			preUserKey, _ := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_key_"+newNodeName)
			preUserDbi, _ := item.DataBagItemMap(preUserKey)

			// add a new node and client that does not match search string.
			Must(i.createClients(newNodeName + "0"))
			Must(i.deleteClients(newNodeName + "1"))

			// perform refresh.
			_, err := i.Service.Refresh(pl)
			sr.assertNoError("refresh skip reencrypt", err)

			// gather post-refresh data.
			postKeys, _ := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_keys")
			postKeysDbi, _ := item.DataBagItemMap(postKeys)
			postUserKey, _ := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_key_"+newNodeName)
			postUserDbi, _ := item.DataBagItemMap(postUserKey)
			sr.assertEqual(fmt.Sprintf("post-refresh key for %s", newNodeName), postUserDbi[newNodeName], preUserDbi[newNodeName])
			sr.assertEqual("keys mode unchanged by payload", preKeysDbi["mode"].(string), postKeysDbi["mode"].(string))
			sr.assertEqual("post-refresh client list", postKeysDbi["clients"], []string{"fakehost1", "testhost1", "testhost10"})

			_, err = i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_key_testhost11")
			sr.assertError("post-refresh unknown client key removed", err)

			_, err = i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_key_testhost10")
			sr.assertNoError("new client sparse key created", err)

			assertSparseKeysOnly(sr, i.Service, vaultName, vaultItemName, postKeysDbi, "admins")
			assertSparseKeysOnly(sr, i.Service, vaultName, vaultItemName, postKeysDbi, "clients")

			return sr
		},
	}
}

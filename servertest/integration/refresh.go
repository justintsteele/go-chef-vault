package integration

import (
	"fmt"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/vault"
)

func (i *IntegrationService) refresh() (result *vault.UpdateResponse, err error) {
	const nodeName = "testhost2"
	defer func() {
		// Delete our new node and client so we can re-run the tests without panics
		if err := i.Service.Client.Nodes.Delete(nodeName); err != nil {
			return
		}

		if err := i.Service.Client.Clients.Delete(nodeName); err != nil {
			return
		}
	}()

	pl := &vault.Payload{
		VaultName:     vaultName,
		VaultItemName: vaultItemName,
		Clean:         true,
		SkipReencrypt: true,
	}

	// Here we add a new node and client so that the search query we added in update has something new to find.
	newNode := chef.NewNode(nodeName)
	node, err := i.Service.Client.Nodes.Post(newNode)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Created new node: %s\n", node.Uri)
	newClient := chef.ApiNewClient{
		Name:       newNode.Name,
		ClientName: newNode.Name,
		Validator:  false,
	}

	client, err := i.Service.Client.Clients.Create(newClient)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Created new client: %s\n", client.Uri)

	result, err = i.Service.Refresh(pl)
	if err != nil {
		return
	}

	// report on client keys here so you can see that it added the new one.
	dbr, dberr := i.Service.Client.DataBags.GetItem(vaultName, vaultItemName+"_keys")
	if dberr != nil {
		return
	}
	report("Get Item Keys:", dbr)

	return
}

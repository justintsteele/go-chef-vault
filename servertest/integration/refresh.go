package integration

import (
	"github.com/justintsteele/go-chef-vault"
)

func (i *IntegrationService) refresh() (result *vault.UpdateResponse, err error) {
	pl := &vault.Payload{
		VaultName:     vaultName,
		VaultItemName: vaultItemName,
		Clean:         true,
		SkipReencrypt: true,
	}

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

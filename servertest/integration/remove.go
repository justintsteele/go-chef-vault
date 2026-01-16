package integration

import (
	"encoding/json"

	"github.com/justintsteele/go-chef-vault"
)

func (i *IntegrationService) remove() (result *vault.UpdateResponse, err error) {
	var raw map[string]interface{}
	vaultItem := `{"fuz": { "buz": "buz-value-5" } }`
	if err = json.Unmarshal([]byte(vaultItem), &raw); err != nil {
		return
	}

	pl := &vault.Payload{
		VaultName:     vaultName,
		VaultItemName: vaultItemName,
		Clients:       []string{fakeNodeName},
		CleanUnknown:  true,
		Content:       raw,
	}

	result, err = i.Service.Remove(pl)
	if err != nil {
		return
	}

	// report on client keys here so you can see that it added the new one.
	dbr, dberr := i.Service.GetItem(vaultName, vaultItemName)
	if dberr != nil {
		return
	}
	report("Get Post-Remove Item:", dbr)

	return
}

package integration

import (
	"github.com/justintsteele/go-chef-vault"
)

func (i *IntegrationService) rotate() (result *vault.RotateResponse, err error) {
	pl := &vault.Payload{
		VaultName:     vaultName,
		VaultItemName: vaultItemName,
		CleanUnknown:  true,
	}

	result, err = i.Service.RotateKeys(pl)
	if err != nil {
		return
	}

	report("Rotate Keys:", result)

	return
}

func (i *IntegrationService) rotateAllKeys() (result []vault.RotateResponse, err error) {
	result, err = i.Service.RotateAllKeys()
	if err != nil {
		return
	}

	report("Rotate All Keys:", result)

	return
}

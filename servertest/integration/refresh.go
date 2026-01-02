package integration

import (
	"github.com/justintsteele/go-chef-vault/vault"
)

func refresh(service *vault.Service) (result *vault.UpdateResponse, err error) {
	pl := &vault.Payload{
		VaultName:     vaultName,
		VaultItemName: vaultItemName,
		Clean:         true,
		SkipReencrypt: true,
	}

	result, err = service.Refresh(pl)
	if err != nil {
		return
	}

	return
}

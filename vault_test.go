package vault

import (
	"testing"

	"github.com/justintsteele/go-chef-vault/item_keys"
	"github.com/stretchr/testify/require"
)

func TestPayload_EffectiveKeysMode_DefaultWhenOmitted(t *testing.T) {
	pl := &Payload{
		VaultName:     "testvault1",
		VaultItemName: "testsecret1",
	}

	keysMode := pl.effectiveKeysMode()
	require.Equal(t, keysMode, item_keys.KeysModeDefault)
}

func TestPayload_EffectiveKeysMode_OverridesModeWhenProvided(t *testing.T) {
	mode := item_keys.KeysModeSparse
	pl := &Payload{
		VaultName:     "testvault1",
		VaultItemName: "testsecret1",
		KeysMode:      &mode,
	}

	keysMode := pl.effectiveKeysMode()
	require.Equal(t, keysMode, item_keys.KeysModeSparse)
}

func TestPayload_ResolveKeysMode_NoChangeWhenOmitted(t *testing.T) {
	pl := &Payload{
		VaultName:     "testvault1",
		VaultItemName: "testsecret1",
	}
	keysMode, state := pl.resolveKeysMode(item_keys.KeysModeSparse)
	require.Equal(t, keysMode, item_keys.KeysModeSparse)
	require.Equal(t, state, &item_keys.KeysModeState{
		Current: item_keys.KeysModeSparse,
		Desired: item_keys.KeysModeSparse,
	})
}

func TestPayload_MergeKeyActors_CleanFalse(t *testing.T) {
	pl := &Payload{
		VaultName:     "testvault1",
		VaultItemName: "testsecret1",
		Admins:        []string{"testadmin2"},
	}
	state := &item_keys.VaultItemKeys{
		Id:      "testsecret1_keys",
		Admins:  []string{"testadmin1"},
		Clients: []string{"testclient1", "testclient2"},
	}

	pl.mergeKeyActors(state)
	require.Equal(t, pl.Admins, []string{"testadmin1", "testadmin2"})
}

func TestPayload_MergeKeyActors_CleanTrue(t *testing.T) {
	pl := &Payload{
		VaultName:     "testvault1",
		VaultItemName: "testsecret1",
		Admins:        []string{"testadmin2"},
		Clean:         true,
	}
	state := &item_keys.VaultItemKeys{
		Id:      "testsecret1_keys",
		Admins:  []string{"testadmin1"},
		Clients: []string{"testclient1", "testclient2"},
	}

	pl.mergeKeyActors(state)
	require.Equal(t, pl.Clients, []string(nil))
}

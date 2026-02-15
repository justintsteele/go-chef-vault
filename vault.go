package vault

import (
	"errors"

	"github.com/justintsteele/go-chef-vault/item_keys"
)

var (
	// ErrNilPayload is returned when a nil *Payload is passed to a public API.
	ErrNilPayload = errors.New("vault: payload cannot be nil")

	// ErrMissingVaultName is returned when VaultName is empty.
	ErrMissingVaultName = errors.New("vault: missing VaultName")

	// ErrMissingVaultItemName is returned when VaultItemName is empty.
	ErrMissingVaultItemName = errors.New("vault: missing VaultItemName")
)

// Payload represents the input parameters used to create, update, or refresh a vault item.
type Payload struct {
	VaultName     string
	VaultItemName string
	Content       map[string]interface{}
	KeysMode      *item_keys.KeysMode
	SearchQuery   *string
	Admins        []string
	Clients       []string
	Clean         bool
	CleanUnknown  bool
	SkipReencrypt bool
}

// validatePayload ensures that required fields are provided in a given payload.
func (p *Payload) validatePayload() error {
	if p == nil {
		return ErrNilPayload
	}

	if p.VaultName == "" {
		return ErrMissingVaultName
	}

	if p.VaultItemName == "" {
		return ErrMissingVaultItemName
	}
	return nil
}

// effectiveKeysMode returns the effective keys mode, defaulting when none is specified.
func (p *Payload) effectiveKeysMode() item_keys.KeysMode {
	if p.KeysMode == nil {
		return item_keys.KeysModeDefault
	}
	return *p.KeysMode
}

// resolveKeysMode determines the effective keys mode and returns the corresponding mode state.
func (p *Payload) resolveKeysMode(current item_keys.KeysMode) (item_keys.KeysMode, *item_keys.KeysModeState) {
	if p.KeysMode == nil {
		return current, &item_keys.KeysModeState{
			Current: current,
			Desired: current,
		}
	}

	return *p.KeysMode, &item_keys.KeysModeState{
		Current: current,
		Desired: *p.KeysMode,
	}
}

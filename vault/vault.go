package vault

import "go-chef-vault/vault/item_keys"

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
	SkipReencrypt bool
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

// mergeKeyActors merges payload actors into the existing key state, honoring the clean flag.
func (p *Payload) mergeKeyActors(state *item_keys.VaultItemKeys) {
	if p.Admins != nil {
		state.Admins = item_keys.MergeClients(state.Admins, p.Admins)
	}

	if p.Clean {
		state.Clients = nil
	}

	if p.Clients != nil {
		if p.Clean {
			state.Clients = p.Clients
		} else {
			state.Clients = item_keys.MergeClients(state.Clients, p.Clients)
		}
	}
}

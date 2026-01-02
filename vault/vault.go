package vault

import "go-chef-vault/vault/item_keys"

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

// effectiveKeysMode allows the payload to not contain a KeysMode and still enact the default
func (p *Payload) effectiveKeysMode() item_keys.KeysMode {
	if p.KeysMode == nil {
		return item_keys.KeysModeDefault
	}
	return *p.KeysMode
}

// resolveKeysMode resolves the desired keys mode from the payload and the current mode
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

// mergeKeyActors merges the actors from the payload into the state, respecting the clean flag
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

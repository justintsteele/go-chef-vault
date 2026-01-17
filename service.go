package vault

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

// Service provides Vault operations backed by a Chef Server client.
type Service struct {
	Client *chef.Client
}

// Response represents the basic structure of a response from a Vault operation.
type Response struct {
	URI string `json:"uri"`
}

// clientSearchResult represents a single row returned from a Chef partial client search.
type clientSearchResult struct {
	Name string `json:"name"`
}

// NewService returns a Service configured with the given Chef client.
func NewService(client *chef.Client) *Service {
	return &Service{
		Client: client,
	}
}

// vaultURL constructs the canonical URL for a vault resource.
func (s *Service) vaultURL(name string) string {
	ref := &url.URL{
		Path: path.Join("data", name),
	}

	return s.Client.BaseURL.ResolveReference(ref).String()
}

// bagIsVault returns bool of whether the specified data bag a vault.
//
// References:
//   - Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_base.rb#L51
func (s *Service) bagIsVault(bagName string) (bool, error) {
	rawItems, err := s.Client.DataBags.ListItems(bagName)
	if err != nil {
		return false, err
	}

	items := *rawItems
	for raw := range items {
		if strings.HasSuffix(raw, "_keys") {
			base := strings.TrimSuffix(raw, "_keys")
			if _, ok := items[base]; ok {
				return true, nil
			}
		}
	}
	return false, nil
}

// bagItemIsEncrypted determines whether the data bag item contains the encrypted_data key of an encrypted data bag.
func (s *Service) bagItemIsEncrypted(vaultName, vaultItem string) (bool, error) {
	dbi, err := s.Client.DataBags.GetItem(vaultName, vaultItem)
	if err != nil {
		return false, err
	}

	dbm, err := item.DataBagItemMap(dbi)
	if err != nil {
		return false, err
	}

	for key, dbv := range dbm {
		if key == "id" {
			continue
		}

		dbim, ok := dbv.(map[string]any)
		if !ok {
			continue
		}

		if _, ok := dbim["encrypted_data"]; ok {
			return true, nil
		}
	}

	return false, nil
}

// getClientsFromSearch returns the names of clients matching the search query.
func (s *Service) getClientsFromSearch(payload *Payload) ([]string, error) {
	if payload.SearchQuery == nil {
		return nil, nil
	}

	plan := item_keys.BuildClientSearchPlan(payload.SearchQuery)

	rows, err := s.executeClientSearch(plan)
	if err != nil {
		return nil, err
	}

	clients := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.Name != "" {
			clients = append(clients, r.Name)
		}
	}
	return clients, nil
}

// executeClientSearch executes a client search plan against the Chef Server and returns the raw results.
func (s *Service) executeClientSearch(plan *item_keys.ClientSearchPlan) ([]clientSearchResult, error) {
	if plan == nil {
		return nil, nil
	}

	result, err := s.Client.Search.PartialExecJSON(plan.Index, plan.Query, plan.Fields)
	if err != nil {
		return nil, err
	}

	rows := make([]clientSearchResult, 0, len(result.Rows))
	for _, row := range result.Rows {
		var r clientSearchResult
		if err := json.Unmarshal(row.Data, &r); err != nil {
			return nil, err
		}
		rows = append(rows, r)
	}

	return rows, nil
}

// loadActorKey retrieves the encrypted shared key for the specified actor.
func (s *Service) loadActorKey(vaultName, vaultItem string) (string, error) {
	rawKeys, err := s.Client.DataBags.GetItem(vaultName, vaultItem+"_keys")
	if err != nil {
		return "", err
	}

	keysMap, err := item.DataBagItemMap(rawKeys)
	if err != nil {
		return "", err
	}

	var actor = s.Client.Auth.ClientName
	var actorKey interface{}
	actorKey, ok := keysMap[actor]
	if !ok {
		// not in default key, trying sparse keys
		rawSparseKey, err := s.Client.DataBags.GetItem(vaultName, vaultItem+"_key_"+actor)
		if err != nil {
			return "", fmt.Errorf("%s/%s is not encrypted with your public key", vaultName, vaultItem)
		}
		sparseKeyMap, err := item.DataBagItemMap(rawSparseKey)
		if err != nil {
			return "", err
		}

		actorKey, ok = sparseKeyMap[actor]
		if !ok {
			return "", fmt.Errorf("%s/%s is not encrypted with your public key", vaultName, vaultItem)
		}

	}
	keyStr, ok := actorKey.(string)
	if !ok {
		return "", fmt.Errorf("%s/%s contains an invalid key format for actor %q", vaultName, vaultItem, actor)
	}
	return keyStr, nil
}

// loadSharedSecret decrypts and returns the vault shared secret using the current actor's private key and encrypted shared key.
func (s *Service) loadSharedSecret(payload *Payload) ([]byte, error) {
	actorKey, err := s.loadActorKey(payload.VaultName, payload.VaultItemName)
	if err != nil {
		return nil, err
	}

	secret, err := item_keys.DecryptSharedSecret(actorKey, s.Client.Auth.PrivateKey)

	if err != nil {
		return nil, fmt.Errorf("unable to decrypt shared secret with available credentials")
	}

	return secret, nil
}

package vault

import (
	"encoding/json"
	"net/url"
	"path"
	"strings"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/item"
	"github.com/justintsteele/go-chef-vault/item_keys"
)

// Service provides Vault operations backed by a Chef Server client.
type Service struct {
	Client    *chef.Client
	authorize func(key string) ([]byte, error)
}

// Response represents the basic structure of a response from a Vault operation.
type Response struct {
	URI string `json:"uri"`
}

// NewService returns a Service configured with the given Chef client.
func NewService(client *chef.Client) *Service {
	vs := &Service{
		Client: client,
	}
	return vs
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
		return []string{}, nil
	}

	plan := item_keys.BuildClientSearchPlan(payload.SearchQuery)

	rows, err := s.executeClientSearch(plan)
	if err != nil {
		return nil, err
	}

	return item_keys.ExtractClients(rows)
}

// executeClientSearch executes a client search plan against the Chef Server and returns the raw results.
func (s *Service) executeClientSearch(plan *item_keys.ClientSearchPlan) ([]json.RawMessage, error) {
	if plan == nil {
		return nil, nil
	}

	result, err := s.Client.Search.PartialExecJSON(plan.Index, plan.Query, plan.Fields)
	if err != nil {
		return nil, err
	}

	rows := make([]json.RawMessage, 0, len(result.Rows))
	for _, row := range result.Rows {
		rows = append(rows, row.Data)
	}

	return rows, nil
}

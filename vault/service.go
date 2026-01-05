package vault

import (
	"encoding/json"
	"go-chef-vault/vault/item_keys"
	"net/url"
	"path"

	"github.com/go-chef/chef"
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

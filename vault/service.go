package vault

import (
	"encoding/json"
	"go-chef-vault/vault/item_keys"
	"net/url"
	"path"

	"github.com/go-chef/chef"
)

type Service struct {
	Client    *chef.Client
	authorize func(key string) ([]byte, error)
}

type Response struct {
	URI string `json:"uri"`
}

// NewService is a constructor for Service. This is used by other vault service methods to authorize access to a vault item.
func NewService(client *chef.Client) *Service {
	vs := &Service{
		Client: client,
	}
	return vs
}

// vaultURL normalizes the vault URL for the response URIs
func (s *Service) vaultURL(name string) string {
	ref := &url.URL{
		Path: path.Join("data", name),
	}

	return s.Client.BaseURL.ResolveReference(ref).String()
}

// deriveItemKey returns the caller's decrypted AES key
func (s *Service) deriveItemKey(encKey string) ([]byte, error) {
	return item_keys.DeriveAESKeyForVault(encKey, s.Client.Auth.PrivateKey)
}

// getClientsFromSearch takes a search query from the vault payload, executes the search and returns a list of client names satisfied by the search
func (s *Service) getClientsFromSearch(payload *Payload) ([]string, error) {
	plan := item_keys.BuildClientSearchPlan(payload.SearchQuery)

	rows, err := s.executeClientSearch(plan)
	if err != nil {
		return nil, err
	}

	return item_keys.ExtractClients(rows)
}

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

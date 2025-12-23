package vault

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/go-chef/chef"
)

type Service struct {
	Client    *chef.Client
	authorize func(key string) ([]byte, error)
}

type VaultResponse struct {
	URI string `json:"uri"`
}

type KeysMode string

const (
	KeysModeDefault KeysMode = "default"
	KeysModeSparse  KeysMode = "sparse"
)

type VaultPayload struct {
	VaultName     string
	VaultItemName string
	Content       map[string]interface{}
	KeysMode      *KeysMode
	SearchQuery   *string
	Admins        []string
	Clients       []string
	Clean         bool
}

func (p *VaultPayload) EffectiveKeysMode() KeysMode {
	if p.KeysMode == nil {
		return KeysModeDefault
	}
	return *p.KeysMode
}

// NewService is a constructor for Service. This is used by other vault service methods to authorize access to a vault item.
func NewService(client *chef.Client) *Service {
	vs := &Service{
		Client: client,
	}
	vs.authorize = vs.authorizeVaultItem
	return vs
}

// authorizeVaultItem validates the requesting client is authorized for the vault and returns the private key
func (v *Service) authorizeVaultItem(key string) ([]byte, error) {
	encKey, err := base64.StdEncoding.DecodeString(cleanB64(key))
	if err != nil {
		return nil, err
	}

	secret, err := rsa.DecryptPKCS1v15(
		rand.Reader,
		v.Client.Auth.PrivateKey,
		encKey,
	)
	if err != nil {
		return nil, fmt.Errorf("error decrypting key: %s", err)
	}

	sum := sha256.Sum256(secret)
	aesKey := sum[:]

	if len(aesKey) != 32 {
		return nil, fmt.Errorf("error decrypting key: invalid length")
	}

	return aesKey, nil
}

// bagIsVault returns bool of whether the specified data bag a vault
//
//	Chef-Vault Source: https://github.com/chef/chef-vault/blob/main/lib/chef/knife/vault_base.rb#L51
func (v *Service) bagIsVault(bagName string) bool {
	rawItems, err := v.Client.DataBags.ListItems(bagName)
	if err != nil {
		return false
	}

	items := *rawItems

	for item := range items {
		if strings.HasSuffix(item, "_keys") {
			base := strings.TrimSuffix(item, "_keys")
			if _, ok := items[base]; ok {
				return true
			}
		}
	}
	return false
}

// getClientsFromSearch takes a search query from the vault payload, executes the search and returns a list of client names satisfied by the search
func (v *Service) getClientsFromSearch(payload *VaultPayload) ([]string, error) {
	part := make(map[string]interface{})
	part["name"] = []string{"name"}
	vault, err := v.Client.Search.PartialExecJSON("node", *payload.SearchQuery, part)
	if err != nil {
		panic(err)
	}

	type clientNames struct {
		Name string `json:"name"`
	}

	names := make([]string, 0, len(vault.Rows))
	for _, node := range vault.Rows {
		var n clientNames
		if err := json.Unmarshal(node.Data, &n); err != nil {
			panic(err)
		}

		if n.Name != "" {
			names = append(names, n.Name)
		}
	}
	return names, nil
}

func (v *Service) vaultURL(name string) string {
	ref := &url.URL{
		Path: path.Join("data", name),
	}

	return v.Client.BaseURL.ResolveReference(ref).String()
}

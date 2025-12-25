package integration

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/go-chef/chef"
	"github.com/justintsteele/go-chef-vault/vault"
)

const (
	goiardiServerURL = "http://localhost:4545"
	goiardAdminUser  = "admin"
	goiardiUser      = "rloggia"
)

func RunGoiardiInit(cfg Config) error {
	Must(cfg.ensureWorkDir())
	Must(cfg.ensureUserConfig(goiardAdminUser))
	Must(cfg.ensureUserConfig(goiardiUser))
	Must(cfg.createChefUser())
	return nil
}

func (c *Config) ensureWorkDir() error {
	if _, err := os.Stat(c.WorkDir); os.IsNotExist(err) {
		return os.MkdirAll(c.WorkDir, 0755)
	}
	return nil
}

func (c *Config) ensureUserConfig(user string) error {
	_, err := os.Stat(fmt.Sprintf("%s/%s.pem", c.WorkDir, user))
	if err != nil {
		if err := generateRSAKeyPair(c.WorkDir, user); err != nil {
			return err
		}
	}

	content := fmt.Sprintf(`current_dir = File.dirname(__FILE__)
node_name "%s"
client_key "#{current_dir}/%s"
chef_server_url "%s"
ssl_verify_mode :verify_none
knife[:vault_mode] = "client"`,
		user,
		fmt.Sprintf("%s.pem", user),
		goiardiServerURL,
	)
	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(c.WorkDir), 0700); err != nil {
		return fmt.Errorf("create WorkDir dir: %w", err)
	}

	if err := os.WriteFile(fmt.Sprintf("%s/%s.rb", c.WorkDir, user), []byte(content), 0600); err != nil {
		return fmt.Errorf("write knife.rb: %w", err)
	}

	return nil
}

func (c *Config) createChefUser() error {
	c.Knife = fmt.Sprintf("%s/%s.rb", c.WorkDir, goiardAdminUser)
	client := c.mustCreateClient()
	_, err := client.Users.Get(goiardiUser)
	if err == nil {
		return err
	}

	pubKey, err := os.ReadFile(fmt.Sprintf("%s/%s.pub", c.WorkDir, goiardiUser))
	if err != nil {
		return err
	}
	user := map[string]any{
		"name":       goiardiUser,
		"email":      "robert.loggia@email.com",
		"password":   "enjoyyourbreakfast",
		"public_key": string(pubKey),
	}
	if err := createUser(user, client, &chef.UserResult{}); err != nil {
		return err
	}

	return nil
}

type userDeleteResponse struct {
	URI string `json:"uri"`
}

func deleteUser(service *vault.Service) (response *userDeleteResponse, err error) {
	if dErr := service.Client.Users.Delete(goiardiUser); err != nil {
		return nil, dErr
	}

	ref := &url.URL{
		Path: path.Join("users", goiardiUser),
	}

	response = &userDeleteResponse{URI: service.Client.BaseURL.ResolveReference(ref).String()}
	return
}

func createUser(user map[string]any, client *chef.Client, v interface{}) error {
	// we have to do this for goiardi because go-chef create user struct has "username" but the version of the API in goiardi expects "name"
	body, err := chef.JSONReader(user)
	if err != nil {
		return err
	}

	req, err := client.NewRequest("POST", "users", body)
	if err != nil {
		return err
	}

	res, err := client.Do(req, v)
	if res != nil {
		defer res.Body.Close()
	}

	return err
}

func generateRSAKeyPair(dir, name string) error {
	// Generate RSA key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate rsa key: %w", err)
	}

	privPath := filepath.Join(dir, name+".pem")
	pubPath := filepath.Join(dir, name+".pub")

	// ----- Private key -----
	privDER := x509.MarshalPKCS1PrivateKey(key)
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	}

	privFile, err := os.OpenFile(privPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open private key file: %w", err)
	}
	defer privFile.Close()

	if err := pem.Encode(privFile, privBlock); err != nil {
		return fmt.Errorf("encode private key: %w", err)
	}

	// ----- Public key -----
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return fmt.Errorf("marshal public key: %w", err)
	}

	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	}

	pubFile, err := os.OpenFile(pubPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open public key file: %w", err)
	}
	defer pubFile.Close()

	if err := pem.Encode(pubFile, pubBlock); err != nil {
		return fmt.Errorf("encode public key: %w", err)
	}

	return nil
}

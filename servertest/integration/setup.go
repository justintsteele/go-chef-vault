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
)

const (
	goiardiStateFile = ".goiardi-servertest.state"
	goiardiServerURL = "http://localhost:4545"
	goiardiAdminUser = "admin"
	goiardiUser      = "rloggia"
)

var requiredFiles = []string{
	goiardiAdminUser + ".rb",
	goiardiAdminUser + ".pem",
	goiardiUser + ".rb",
	goiardiUser + ".pem",
}

func RunGoiardiInit(cfg Config) error {
	Must(cfg.PrepareSandbox())
	Must(cfg.createChefUser())
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
	if err := os.MkdirAll(c.WorkDir, 0700); err != nil {
		return err
	}

	if err := os.WriteFile(fmt.Sprintf("%s/%s.rb", c.WorkDir, user), []byte(content), 0600); err != nil {
		return fmt.Errorf("write knife.rb: %w", err)
	}

	return nil
}

func (c *Config) createChefUser() error {
	c.Knife = fmt.Sprintf("%s/%s.rb", c.WorkDir, goiardiAdminUser)
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

func (c *Config) PrepareSandbox() error {
	if c.Target != TargetGoiardi {
		return nil
	}

	if c.isSandboxBootstrapped() {
		if err := c.validateSandbox(); err != nil {
			return fmt.Errorf(
				"existing sandbox invalid (%v); delete %s and re-run",
				err,
				c.State,
			)
		}
		return nil
	}

	Must(c.bootstrapSandbox())
	return nil
}

func (c *Config) isSandboxBootstrapped() bool {
	if _, err := os.Stat(c.State); err == nil {
		return true
	}
	return false
}

func (c *Config) bootstrapSandbox() error {
	if err := os.MkdirAll(c.WorkDir, 0700); err != nil {
		return err
	}

	Must(c.ensureUserConfig(goiardiAdminUser))
	Must(c.ensureUserConfig(goiardiUser))

	return os.WriteFile(
		c.State,
		[]byte(c.WorkDir),
		0600,
	)
}

func (c *Config) validateSandbox() error {
	for _, rel := range requiredFiles {
		reqFile := filepath.Join(c.WorkDir, rel)
		info, err := os.Stat(reqFile)
		if err != nil {
			return fmt.Errorf("missing %s", rel)
		}
		if info.IsDir() {
			return fmt.Errorf("expected file, found directory: %s", rel)
		}
	}
	return nil
}

func (c *Config) DestroySandbox() error {
	entries, err := os.ReadDir(c.WorkDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		file := filepath.Join(c.WorkDir, e.Name())
		if err := os.RemoveAll(file); err != nil {
			return err
		}
	}

	if err := os.RemoveAll(c.WorkDir); err != nil {
		return err
	}

	if err := os.Remove(c.State); err != nil {
		return err
	}

	return nil
}

type userDeleteResponse struct {
	URI string `json:"uri"`
}

func (i *IntegrationService) createClients(name string) error {
	newNode := chef.NewNode(name)
	node, err := i.Service.Client.Nodes.Post(newNode)
	if err != nil {
		return err
	}
	fmt.Printf("Created new node: %s\n", node.Uri)
	newClient := chef.ApiNewClient{
		Name:       newNode.Name,
		ClientName: newNode.Name,
		Validator:  false,
	}
	client, err := i.Service.Client.Clients.Create(newClient)
	if err != nil {
		return err
	}
	fmt.Printf("Created new client: %s\n", client.Uri)

	return nil
}

func (i *IntegrationService) deleteClients(name string) error {
	if err := i.Service.Client.Nodes.Delete(name); err != nil {
		return err
	}
	fmt.Printf("Deleted node: %s\n", name)

	if err := i.Service.Client.Clients.Delete(name); err != nil {
		return err
	}
	fmt.Printf("Deleted client: %s\n", name)
	return nil
}

func (i *IntegrationService) deleteUser() (response *userDeleteResponse, err error) {
	if dErr := i.Service.Client.Users.Delete(goiardiUser); err != nil {
		return nil, dErr
	}

	ref := &url.URL{
		Path: path.Join("users", goiardiUser),
	}

	response = &userDeleteResponse{URI: i.Service.Client.BaseURL.ResolveReference(ref).String()}
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
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Printf("failed to close response body")
		}
	}()

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
	defer func() {
		if err := privFile.Close(); err != nil {
			fmt.Printf("failed to close private key file")
		}
	}()

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
	defer func() {
		if err := pubFile.Close(); err != nil {
			fmt.Printf("failed to close public key file")
		}
	}()

	if err := pem.Encode(pubFile, pubBlock); err != nil {
		return fmt.Errorf("encode public key: %w", err)
	}

	return nil
}

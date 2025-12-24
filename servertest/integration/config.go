package integration

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/go-chef/chef"
)

type Target int

const (
	TargetGoiardi Target = iota
	TargetChefServer
)

type Config struct {
	Target  Target
	Keep    bool
	WorkDir string
	Knife   string
	Admin   string
	User    string
}

func LoadConfig() Config {
	target := flag.String("target", "chefserver", "goiardi or chefserver")
	knife := flag.String("knife", filepath.Join(os.Getenv("HOME"), ".chef"), "path to knife.rb (chefserver only)")
	keep := flag.Bool("keep-workdir", false, "keep goiardi bootstrap workdir")

	flag.Parse()

	cfg := Config{
		WorkDir: filepath.Dir(*knife),
		Keep:    *keep,
	}

	switch *target {
	case "goiardi":
		cfg.Target = TargetGoiardi
		cfg.WorkDir = "servertest/.chef"
		cfg.Admin = goiardAdminUser
		cfg.User = goiardiUser
	case "chefserver":
		cfg.Target = TargetChefServer
	default:
		log.Fatalf("unknown target: %s", *target)
	}

	return cfg
}

func (c *Config) loadKnifeConfig() (*chef.Config, error) {
	data, err := os.ReadFile(c.Knife)
	if err != nil {
		panic(err)
	}

	clientRb, err := chef.NewClientRb(string(data), c.WorkDir)
	if err != nil {
		panic(err)
	}
	return &chef.Config{
		Name:                  clientRb.NodeName,
		Key:                   clientRb.ClientKey,
		BaseURL:               clientRb.ChefServerUrl + "/",
		AuthenticationVersion: "1.0",
	}, nil
}

func (c *Config) mustCreateClient() *chef.Client {
	knife, err := c.loadKnifeConfig()
	if err != nil {
		panic(err)
	}

	client, err := chef.NewClient(knife)
	if err != nil {
		panic(err)
	}

	return client
}

package integration

import (
	"flag"
	"fmt"
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
	Target Target
	Init   bool

	KnifeRb string
	WorkDir string
}

func LoadConfig() Config {
	init := flag.Bool("init", false, "initialize goiardi")
	target := flag.String("target", "goiardi", "goiardi or chefserver")
	knife := flag.String("knife", filepath.Join(os.Getenv("HOME"), ".chef"), "path to knife.rb (chefserver only)")

	flag.Parse()

	cfg := Config{
		Init:    *init,
		KnifeRb: *knife,
		WorkDir: ".servertest",
	}

	switch *target {
	case "goiardi":
		cfg.Target = TargetGoiardi
	case "chefserver":
		cfg.Target = TargetChefServer
	default:
		log.Fatalf("unknown target: %s", *target)
	}

	if cfg.Target == TargetChefServer && cfg.Init {
		log.Fatal("--init is not valid for chefserver")
	}

	return cfg
}

func newClient(cfg *chef.Config) (*chef.Client, error) {
	return chef.NewClient(cfg)
}

func loadKnifeConfig(knifeRb string) (*chef.Config, error) {
	data, err := os.ReadFile(fmt.Sprintf("%s/knife.rb", knifeRb))
	if err != nil {
		panic(err)
	}

	clientRb, err := chef.NewClientRb(string(data), knifeRb)
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

func mustCreateClient(cfg Config) *chef.Client {
	knife, err := loadKnifeConfig(cfg.KnifeRb)
	if err != nil {
		panic(err)
	}

	client, err := chef.NewClient(knife)
	if err != nil {
		panic(err)
	}

	return client
}

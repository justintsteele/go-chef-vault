package integration

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	State   string
	WorkDir string
	Knife   string
	Admin   string
	User    string
}

func LoadConfig() Config {
	target := flag.String("target", "goiardi", "goiardi or chefserver")
	knife := flag.String("knife", filepath.Join(os.Getenv("HOME"), ".chef", "knife.rb"), "path to knife.rb")
	keep := flag.Bool("keep-workdir", false, "keep goiardi integration workdir")

	flag.Parse()

	switch *target {
	case "goiardi":
		var workDir string
		statePath := filepath.Join(os.TempDir(), goiardiStateFile)

		if data, err := os.ReadFile(statePath); err == nil {
			workDir = strings.TrimSpace(string(data))
		} else {
			workDir, err = os.MkdirTemp("", "goiardi-servertest-")
			if err != nil {
				log.Fatalf("unable to create workdir: %v", err)
			}
		}

		return Config{
			Target:  TargetGoiardi,
			State:   filepath.Join(os.TempDir(), goiardiStateFile),
			WorkDir: workDir,
			Keep:    *keep,
		}
	case "chefserver":
		knifePath, err := filepath.Abs(*knife)
		if err != nil {
			log.Fatalf("unable to determine absolute path to knife.rb")
		}

		knifeRb, err := os.Stat(knifePath)
		if err != nil {
			log.Fatalf("cannot access --knife file %s: %v", knifePath, err)
		}

		if knifeRb.IsDir() {
			log.Fatalf("--knife must point to a file, got directory: %s", knifePath)
		}

		if !*keep {
			log.Printf("WARNING: --keep-workdir ignored for chefserver target (cleanup disabled)")
		}

		return Config{
			Target: TargetChefServer,
			Knife:  knifePath,
			Keep:   true,
		}
	default:
		log.Fatalf("unknown target: %s", *target)
	}
	return Config{}
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

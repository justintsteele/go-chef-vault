package main

import (
	"log"
	"os"
	"servertest/integration"
)

func main() {
	cfg := integration.LoadConfig()

	defer func() {
		if !cfg.Keep {
			if err := os.RemoveAll(cfg.WorkDir); err != nil {
				log.Printf("warning: failed to clean workdir: %v", err)
			}
		}
	}()

	if cfg.Target == integration.TargetGoiardi {
		integration.Must(integration.RunGoiardiInit(cfg))
	}

	integration.Must(integration.RunVault(cfg))
}

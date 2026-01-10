package main

import (
	"log"
	"servertest/integration"
)

func main() {
	cfg := integration.LoadConfig()

	defer func() {
		if !cfg.Keep {
			if err := cfg.DestroySandbox(); err != nil {
				log.Printf("warning: failed to clean workdir: %v", err)
			}
		}
	}()

	if cfg.Target == integration.TargetGoiardi {
		integration.Must(integration.RunGoiardiInit(cfg))
	}

	integration.Must(integration.RunVault(cfg))
}

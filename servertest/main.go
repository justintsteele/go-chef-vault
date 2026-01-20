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
			if err := cfg.DestroySandbox(); err != nil {
				log.Printf("warning: failed to clean workdir: %v", err)
			}
		}
	}()

	if cfg.Target == integration.TargetGoiardi {
		integration.Must(integration.RunGoiardiInit(cfg))
	}

	reporter := &integration.ConsoleReporter{}
	results, err := integration.RunScenarios(cfg, reporter)
	if err != nil {
		log.Printf("integration run failed: %v", err)
		os.Exit(1)
	}

	if integration.AnyFailed(results) {
		os.Exit(1)
	}

	os.Exit(0)
}

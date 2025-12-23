package main

import (
	"servertest/integration"
)

func main() {
	cfg := integration.LoadConfig()

	if cfg.Target == integration.TargetGoiardi && cfg.Init {
		integration.Must(integration.RunGoiardiInit(cfg))
	}

	integration.Must(integration.RunVault(cfg))
}

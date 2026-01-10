# Chef-Vault API for Chef Server in Golang

A Go implementation of the Chef-Vault API, built as a sidecar to `go-chef` and designed to closely mirror the behavior and semantics of the Ruby `chef-vault` gem.

## Install

```bash
> go get github.com/justintsteele/go-chef-vault@latest
```

## Test

```bash
> go test -v github.com/justintsteele/go-chef-vault/vault
> go run -v github.com/justintsteele/go-chef-vault/servertest --target [chefserver|goiardi]
```

See `servertest/README.md` for further setup and configuration steps for the integration suite.

The integration test suite exercises the vault API against a live Chef Server or goiardi instance to ensure behavioral parity with the Ruby implementation.


## Usage

All public API endpoints require a configured Chef client. There are several ways to accomplish this as documented in [go-chef](https://github.com/go-chef/chef).

Following the example laid out in `servertest` in this project, we start with a knife.rb, read that in and create the client.

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-chef/chef"
	vault "github.com/justintsteele/go-chef-vault"
)

func main() {
	data, err := os.ReadFile("/path/to/.chef/knife.rb")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	clientRb, err := chef.NewClientRb(string(data), "/path/to/.chef")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	knife := &chef.Config{
		Name:                  clientRb.NodeName,
		Key:                   clientRb.ClientKey,
		BaseURL:               clientRb.ChefServerUrl + "/",
		AuthenticationVersion: "1.0",
	}

	client, err := chef.NewClient(knife)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	vaultService := vault.NewService(client)

	result, err := vaultService.GetItem("vault_name", "secret_name")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	jsonRes, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("error marshaling result: %v\n", err)
		return
	}

	fmt.Println(string(jsonRes))
}
```

## Contributing

If you feel like contributing, great! Just fork the repo, make your improvements, and submit a pull request. When adding features, please ensure behavior remains compatible with the Ruby chef-vault gem unless explicitly documented otherwise.
Please also ensure all new features and bug fixes are covered by a unit test. Bonus points if you are able to extend coverage into the `servertest` integration suite. Ensure code passes the lint standards provided by `golangci-lint`.

## Copyright

Copyright 2025-2026, Justin Steele

## License

go-chef-vault is licensed under the Apache 2.0 License. See the LICENSE file for details.
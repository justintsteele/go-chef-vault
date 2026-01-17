# Chef-Vault API for Chef Server in Golang

A Go implementation of the Chef-Vault API, built as a sidecar to `go-chef` and designed to closely mirror the behavior and semantics of the Ruby `chef-vault` gem.

## Install

```bash
> go get github.com/justintsteele/go-chef-vault@latest
```

## API

This library exposes a small, opinionated API that mirrors the behavior of
the Ruby `chef-vault` knife commands. The primary entry point is the `vault.Service`
type, which operates on Chef Vaults and Vault Items.

### Core Types

- `vault.Service`  
  The main client used to perform vault operations. It wraps a `go-chef` client
  and applies Chef-Vault semantics on top of it.

- `vault.Payload`  
  A request structure used by mutating operations (`Create`, `Update`, `Rotate`,
  `Remove`). Fields are optional unless required by the operation, matching
  Ruby Chef-Vault behavior.

### Read Operations

- `GetItem(vaultName, itemName string)`  
  Retrieves and decrypts a vault item, returning the plaintext data.

- `List()`
  Retrieves a list of all vaults in the Chef Server.

- `ListItems(name string)`
  Retrieves a list of all items in a specified vault.

- `IsVault(vaultName string)`  
  Determines whether a data bag represents a Chef Vault.

- `ItemType(vaultName, vaultItem string)`  
  Determines whether the data bag item is a vault, encrypted data bag, or a normal data bag item.

### Write / Mutating Operations

- `Create(payload *Payload)`  
  Creates a new vault and encrypted item.

- `Update(payload *Payload)`  
  Updates vault contents while preserving omitted invariants (key mode,
  search query, existing actors).

- `Delete(payload *Payload)`
  Destroys the entire vault, all the items, and keys from the Chef Server.

- `DeleteItem(payload *Payload)`
  Destroys a specified vault item and its keys 

- `RotateKeys(payload *Payload)`  
  Re-encrypts vault data and keys.

- `RotateAllKeys()`
  Re-encrypts all vaults in the chef servers.

- `Refresh(payload *Payload)`
  Reprocesses the vault search query and ensures all matching nodes have an encrypted secret,
  without modifying existing vault content or access rules.

- `Remove(payload *Payload)`
  Removes data or actors from an existing vault.

### Error Handling

Errors returned by this library may wrap underlying `go-chef` errors.
Helper functions are provided in the `cheferr` package to inspect error
types:

- `cheferr.IsNotFound(err)`
- `cheferr.IsConflict(err)`
- `cheferr.AsChefError(err)`

These helpers are recommended instead of direct type assertions.

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
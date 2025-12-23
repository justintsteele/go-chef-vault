# Chef-Vault API for Chef Server in Golang
This library is written as a sidecar to `go-chef` and is for the purpose of creating, modifying, deleting, and otherwise managing Chef-Vault items in Chef Server.

## Install

    go get github.com/justintsteele/go-chef-vault

## Test

    go get -t github.com/justintsteele/go-chef-vault
    go test -v github.com/justintsteele/go-chef-vault/vault
    go run -v github.com/justintsteele/go-chef-vault/servertest --target [chefserver|goiardi]

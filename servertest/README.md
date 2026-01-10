# go-chef-vault Integration Testing Suite

The `servertest` package is an integration test suite designed to provide real-world examples of how to use the `go-chef-vault` APIs. The package can be configured to execute directly against a local [goiardi](https://github.com/ctdk/goiardi) instance or against a live Chef Server.

## Goiardi

There are installation instructions for getting goiardi running on your machine [here](https://goiardi.readthedocs.io/en/latest/installation.html), or you can follow along below!

### Install and Launch goiardi

Ensure that `go` is in your path. This will install to `~/go/bin` or wherever your `$GOPATH/bin` (or `$GOBIN`) is located.

```bash
> go install github.com/ctdk/goiardi@latest
```

Create a `goiardi.cfg` file.

```ini
listen = "127.0.0.1:4545"
log-level = "debug"
use-auth = false
```

Launch `goiardi` with the config file.

```bash
> ~/go/bin/goiardi -c /path/to/goiardi.cfg &
```

### Run
From the root of the `go-chef-vault` repository, run the `servertest` package:

```bash
> go run servertest 
```

## Chef Server

If the real world is more your speed, and you have a knife.rb, you can use `servertest` for this as well. If your knife.rb is in the default location (`$HOME/.chef/knife.rb`), just pass the `--target chefserver` option. You can override the location of the `knife.rb` with the `--knife /path/to/my/knife.rb` option.

> ⚠️ **Warning**
>
> The integration test suite is intentionally stateful and destructive. 
> These tests will create and delete clients, nodes, and vault items on the target Chef Server. 
> Use only against local or dedicated test environments.

### Run
From the root of the `go-chef-vault` repository, run the `servertest` package:

```bash
> go run servertest --target chefserver
```

## Options

The `servertest` package does provide a few command-line options that can be helpful in debugging or redirecting the default behavior toward a live Chef Server.

```bash
> go run servertest --help
Usage of servertest:
  -keep-workdir
        prevents deletion of the test user, generated knife config, pem files, and sandbox directory
  -knife string
        path to knife.rb (default "$HOME/.chef/knife.rb")
  -target string
        goiardi or chefserver (default "goiardi")

```

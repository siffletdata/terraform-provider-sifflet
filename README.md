# Terraform provider for Sifflet

This is a Terraform provider to manage [Sifflet](https://www.siffletdata.com) resources.

* Sifflet website: https://www.siffletdata.com
* Sifflet documentation: https://docs.siffletdata.com
* Provider documentation: https://registry.terraform.io/providers/Siffletapp/sifflet/latest/docs

Sifflet is the leading end-to-end data observability platform built for data engineers and data consumers. The platform includes data quality monitoring, metadata management, and a data catalog with deep lineage capabilities.

## Project status and support

**This is not an official Sifflet product**.

Sifflet does not provide official support for this provider. This provider should be considered alpha-quality
software.

This provider relies on alpha Sifflet APIs. These APIs may be subject to change without notice.

## Usage

See https://registry.terraform.io/providers/Siffletapp/sifflet/latest/docs.

## Development

### Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.22

### Building The provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

### Adding dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

### Developing the provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To update the generated documentation and code, run `go generate ./...`.

### Local setup

Update your `.terraformrc` like this:

```
provider_installation {

  dev_overrides {
    # Example GOBIN path, will need to be replaced with your own GOBIN path. Default is $GOPATH/bin
    "hashicorp.com/edu/sifflet" = "/Users/<Username>/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

Create a terraform file like this:

```
terraform {
  required_providers {
    sifflet = {
      source = "hashicorp.com/edu/sifflet"
    }
  }
}

provider "sifflet" { }
```

Export the required environment variables:

```
export SIFFLET_HOST="http://localhost:8000" # or your Sifflet API endpoint e.g https://yourinstance.siffletdata.com/api
export SIFFLET_TOKEN="your-api-token"
```


Tip: you can set the Terraform log level to DEBUG with this environment variable:
```
export TF_LOG=DEBUG
```

### Run tests

Ensure the `SIFFLET_TOKEN` environment variable is set, then run this command, subsituting your Sifflet API
endpoint:

```
SIFFLET_HOST="https://yourinstance.siffletdata.com/api" TF_ACC=1 go test -v ./...
```

**Important**: tests create and delete resources in your Sifflet instance. When tests fail or are interrupted, they can leave
dangling resources behind them. Avoid using a production instance for testing.

### Run linters

Install golangci-lint, then run

```
golangci-lint run
```

### Regenerate the Sifflet API client

You can fetch the latest OpenAPI schema from https://docs.siffletdata.com/openapi/. Store it under
``internal/client/openapi.yaml``, then run:

```
go generate ./internal/client
```

**Alpha APIs and known issues**

The `internal/alphaclient` package contains a generated client against 'alpha APIs' - private Sifflet APIs
  subject to chance without notice. Don't use this package for new resources.

Some existing resources are implemented against this client. They will be deprecated or migrated to the stable
  client in the future.

This "alpha" client has the following known issues:

- The JSON lib doesn't support epoch as Time format, you need to replace all `*time.Time` by `*int64`
- The `Update` method is not yet implemented in the OpenAPI schema for the `datasource` resource.
- The `timezoneData` field (used to create a datasource content) is not validated by the API.
- The `/ui/v1/datasources/{id}` endpoint returns nothing on 404 responses.

### Pre-commit hooks

To catch early some common issues that would be rejected in CI, you can install pre-commit hooks:

1. Install [pre-commit](https://pre-commit.com/#install)
1. Run:

```
pre-commit install
```

## Thanks

This provider was contributed by [Benjamin Berriot](https://github.com/IIBenII). Many thanks from the Sifflet
team for your work on this project!

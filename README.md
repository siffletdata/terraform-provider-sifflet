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
- [Go](https://golang.org/doc/install) >= 1.20

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

To generate or update documentation, run `go generate`.

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

provider "sifflet" {
  host  = "http://localhost:8080" # or Sifflet API endpoint
  token = "<SIFFLET_TOKEN>"
}
```

You can use the environment variable `SIFFLET_TOKEN` to set the Sifflet API token instead of hardcoding it in the provider block.

Tip: you can set the Terraform log level to DEBUG with this environment variable:
```
export TF_LOG=DEBUG
```


### Regenerate the Sifflet API client

You can fetch the latest OpenAPI schema from https://docs.siffletdata.com/openapi/. Store it under
``openapi.yaml``, then run:

```
cd internal/alphaclient
go generate
```

**Notes and known issues**

- The JSON lib doesn't support epoch as Time format, you need to replace all `*time.Time` by `*int64`
- The `Update` method is not yet implemented in the OpenAPI schema for the `datasource` resource.
- The `timezoneData` field (used to create a datasource content) is not validated by the API.
- The `/ui/v1/datasources/{id}` endpoint returns nothing on 404 responses.

## Thanks

This provider was contributed by [Benjamin Berriot](https://github.com/IIBenII). Many thanks from the Sifflet
team for your work on this project!

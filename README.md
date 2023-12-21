# Local setup

```
git clone
cd terraform-provider-sifflet
go mod tidy
go mod install .
```

Then update your `.terraformrc` like this:

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

You create a terraform file like this:

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

You can use environment variable `SIFFLET_TOKEN` to set Sifflet token instead of hardcoding in file.

You can set Terraform log as DEBUG with this env variable:
```
export TF_LOG=DEBUG
```


# Regenerate Sifflet client

```
oapi-codegen -package example openapi.yaml > internal/client/sifflet.gen.go
```

**Note**

JSON lib doesn't support epoch as Time format, you need to replace all `*time.Time` by `*int64`
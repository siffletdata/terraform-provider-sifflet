# Terraform provider for Sifflet

* Sifflet website: https://www.siffletdata.com
* Sifflet documentation: https://docs.siffletdata.com
* Provider documentation: https://registry.terraform.io/providers/Siffletdata/sifflet/latest/docs

Sifflet is the leading end-to-end data observability platform built for data engineers and data consumers. The platform includes data quality monitoring, metadata management, and a data catalog with deep lineage capabilities.

## Usage

```hcl
provider "sifflet" {
  host  = "https://example.siffletdata.com/api"
  token = "123azert"
}
```

where:

* `host` is the URL of your Sifflet instance.
    - If you're using a Sifflet SaaS instance: if you access the Sifflet web application `https://example.siffletdata.com`, your API URL is `https://example.siffletdata.com/api`
    - If you're using a self-hosted Sifflet instance, contact your administrator to get the API URL.

* `token` is your Sifflet API token. You can generate a token from the [Sifflet web application](https://docs.siffletdata.com/docs/access-tokens). You can also use the `SIFFLET_TOKEN` environment variable to avoid hardcoding a secret in your configuration.

### Environment variables

The provider reads the following environment variables:

* `SIFFLET_HOST`: Sifflet API hostname. Can also be provided in the provider configuration.
* `SIFFLET_TOKEN`: Sifflet API token. Can also be provided in the provider configuration.
* `TF_APPEND_USER_AGENT`: If set, append this string to the User-Agent header of HTTP requests made by the provider.

Standard Terraform environment variables, such as `TF_LOG`, are also supported.

## Project status and support

This is an official Sifflet project.

This provider should be considered beta-quality software. Please provide feedback and report bugs to Sifflet support, or on the provider [GitHub repository](https://github.com/siffletdata/terraform-provider-sifflet).

With the exception of resources explicitly marked as deprecated, all resources and data sources are considered stable and no breaking changes are expected until the first major release.

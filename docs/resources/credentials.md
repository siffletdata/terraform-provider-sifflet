---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sifflet_credentials Resource - terraform-provider-sifflet"
subcategory: ""
description: |-
  Credentials are used to store secret source connection information, such as usernames, passwords, service account keys, or API tokens.
---

# sifflet_credentials (Resource)

Credentials are used to store secret source connection information, such as usernames, passwords, service account keys, or API tokens.

## Example Usage

```terraform
resource "sifflet_credentials" "example" {
  name        = "credential-name"
  description = "Credential description."
  # Due to API limitations, Terraform can't detect changes to the value that are made outside of Terraform.
  value = "example"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the credentials. Must start and end with a letter, and contain only letters, digits and hyphens. Must be unique in the Sifflet instance.

### Optional

- `description` (String) The description of the credentials.
- `value` (String, Sensitive) The value of the credentials. Due to API limitations, Terraform can't detect changes to this value made outside of Terraform. Mandatory when asking Terraform to create the resource; otherwise, if the resource is imported or was created during a previous apply, this value is optional.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
# Due to API limitations, Terraform can't detect changes on imported credentials.
# Importing credentials will always generate a diff during the first apply, even
# if the configured value is the same as the imported one.
terraform import sifflet_credentials.example 'credentialname'
```

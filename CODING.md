# Coding conventions and general architecture notes

Also see `README.md` for development instructions. This document is intended to help contributors understand
the code base.

## Terraform provider framework

This provider uses the [Terraform provider framework](https://developer.hashicorp.com/terraform/plugin/framework).
Before working on it, it's recommended to read the "Key Concepts" documentation, read the tutorial, and in general
spend some time getting familiar with the framework.

## Code structure

The main entry point for the provider is `internal/provider.go`.

Resources and data sources are defined in packages under `internal/provider/`.
These packages (by convention) should export `Resources()` and `DataSources()` functions that returns the list
of resources and data sources supported by the package.

## Acceptance tests

Acceptance tests for each resource or data source should be implemented in the same folder as the resource or
data source. The test should be in a file named `*_test.go`.
These files should belong to a `*_test` package to avoid dependency cycles (see existing tests for examples).

### Sweepers

Sweepers are used to delete resources left over by any acceptance test (from any test session).

To implement a new sweeper:
* create a file named `sweepers_test.go` in the module exercising the resources to clean
* add a `TestMain` function to this file (see existing files for examples)
* add a `init` function to the acceptance test file (see existing files for examples)

See `README.md` for usage.

## Models, DTOs and Terraform types

The code manipulates three concepts everywhere:
* models are a Go struct representation of a Terraform state (or plan). To make the code easier to follow,
  models should implement (whenever possible) the interfaces defined in the `model` package. See the
  documentation of this package and the related interfaces (`CreatableModel`, `UpdatableModel`, etc.) for
  details. By using this `model` package, the code will naturally follow the same patterns everywhere and be
  easier to understand.
* DTOs (Data Transfer Objects) represent the body of a request or response to the Sifflet API. Models define
  methods that allow to convert them to and from DTOs. DTOs are automatically generated from the API schema
  (see the `client` package for details).
* Terraform types (in the `types` package provided by the plugin framework) should only be used as temporary
  variables inside functions. Avoid method signatures that use them as arguments or return values, prefer
  using models instead.

## Resource example

The `sifflet_user` resource is a simple resource that showcases the basic patterns used in the provider.
See `internal/provider/user`.

## Misc

### Avoid client side validation

In order not to have to update the provider on every change of the API, avoid client-side validation by
default. There are exceptions to this rule:
* easy to make mistakes, when the API expects surprising values. For instance, credential names (must match a
  complicated regex that we validate client-side) or requiring permissions on at least one domain to be able
  to create a user.
* errors that would be time-consuming (e.g in a typical configuration, it would take several minutes for the
  API to report an error)

### Considerations for future work

* This provider contains a lot of boilerplate code. When patterns are better identified, we'll consider
  relying more on code generation.
* This provider relies on unstable APIs (and the corresponding client). Don't use them for new resources.
  We'll remove all related code at some point.

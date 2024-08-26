package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// Use environment variables to configure the provider during tests (SIFFLET_HOST and SIFFLET_TOKEN, see README.md).
	providerConfig = `
provider "sifflet" { }
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"sifflet": providerserver.NewProtocol6WithError(New("test")()),
	}
)

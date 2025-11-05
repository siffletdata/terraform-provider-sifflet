package domain_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"
)

func TestAccDomainDataSource(t *testing.T) {
	// This domain exists on all Sifflet instances.
	domainId := "aaaabbbb-aaaa-bbbb-aaaa-bbbbaaaabbbb"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				data "sifflet_domain" "test" {
					id = "%s"
				}`, domainId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sifflet_domain.test", "name", "All"),
					resource.TestCheckResourceAttr("data.sifflet_domain.test", "description", "Global domain"),
					resource.TestCheckResourceAttr("data.sifflet_domain.test", "id", domainId),
				),
			},
		},
	})
}

func TestAccDomainDataSourceError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + `
				data "sifflet_domain" "test" {
					id = "00000000-0000-0000-0000-000000000000"
				}`,
				ExpectError: regexp.MustCompile("HTTP status code: 404. Details: Domain not found"),
			},
		},
	})
}

package provider

import (
	"fmt"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSourcesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providertests.ProviderConfig() + fmt.Sprintf(`
				data "sifflet_sources" "test" {
					filter = {
						text_search = "test"
						types = ["MYSQL"]
						tags = [ { name = "test" } ]
					}
				}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.sifflet_sources.test", "filter.text_search", "test"),
				),
			},
		},
	})
}

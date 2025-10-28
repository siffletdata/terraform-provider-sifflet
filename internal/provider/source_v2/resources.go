package source_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newSourceV2Resource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newSourceV2SchemasDataSource,
	}
}

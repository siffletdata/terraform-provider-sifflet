package datasource

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func Resources() []func() resource.Resource {
	return []func() resource.Resource{
		newDataSourceResource,
	}
}

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newDatasourcesDataSource,
	}
}

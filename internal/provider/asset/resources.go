package asset

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func DataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newAssetDataSource,
		newAssetsDataSource,
	}
}

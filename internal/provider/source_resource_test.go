package provider

import (
	"fmt"
	"regexp"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func randomSourceName() string {
	return providertests.RandomName()
}

func baseConfig(credName string) string {
	return providertests.ProviderConfig() + fmt.Sprintf(`
    resource "sifflet_credential" "test" {
		name = "%s"
		description = "Description"
		value = "Value"
	}
	`, credName)
}

// BigQuery sources are also used for testing specific attributes and behaviours of the sifflet_source resource.
// For other source types, we only do a simple create/destroy test.
func TestAccBigQuerySource(t *testing.T) {
	sourceName := randomSourceName()
	projectId := providertests.RandomName()
	credName := providertests.RandomCredentialName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								bigquery = {
									project_id = "%s"
									dataset_id = "dataset"
									billing_project_id = "dataset"
								}
							}
						}
						`, sourceName, projectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "bigquery"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.bigquery.project_id", projectId),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "Another description"
							credential = sifflet_credential.test.name
							parameters = {
								bigquery = {
									project_id = "%s"
									dataset_id = "dataset"
									billing_project_id = "dataset"
								}
							}
							timezone = "UTC+1"
						}
						`, sourceName, projectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "description", "Another description"),
				),
			},
		},
	})
}

func TestAccSourceInvalidConfig(t *testing.T) {
	sourceName := randomSourceName()
	projectId := providertests.RandomName()
	database := providertests.RandomName()
	credName := providertests.RandomCredentialName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								bigquery = {
									project_id = "project_id"
									dataset_id = "dataset"
								}
								databricks = {
									catalog = "catalog"
									host = "host"
									http_path = "http_path"
									port = 32
									schema = "schema"
								}
							}
						}
						`, sourceName),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							// No credentials, even though it's required by BigQuery datasources
							parameters = {
								bigquery = {
									project_id = "%s"
									dataset_id = "dataset"
									billing_project_id = "dataset"
								}
							}
						}
						`, sourceName, projectId),
				ExpectError: regexp.MustCompile("Credential is required for this source type, but got an empty string"),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							// Credentials, even if they are not required for Athena datasources
							credential = sifflet_credential.test.name
							parameters = {
								athena = {
									database = "%s"
									datasource = "datasource"
									region = "region"
									role_arn = "arn:aws:iam::123456789012:role/role"
									s3_output_location = "s3://mybucket"
									workgroup = "workgroup"
								}
							}
						}
						`, sourceName, database),

				ExpectError: regexp.MustCompile("Credential is not required for this source type and would be ignored"),
			},
		},
	})
}

func TestAccSourceTags(t *testing.T) {
	sourceName := randomSourceName()
	projectId := providertests.RandomName()
	tagName := providertests.RandomName()
	credName := providertests.RandomCredentialName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								bigquery = {
									project_id = "%s"
									dataset_id = "dataset"
									billing_project_id = "dataset"
								}
							}
						}
						`, sourceName, projectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "tags.#", "0"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								bigquery = {
									project_id = "%s"
									dataset_id = "dataset"
									billing_project_id = "dataset"
								}
							}
							tags = []
						}
						`, sourceName, projectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "tags.#", "0"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_tag" "test" {
							name = "%s"
						}

						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								bigquery = {
									project_id = "%s"
									dataset_id = "dataset"
									billing_project_id = "dataset"
								}
							}
							tags = [{
								name = sifflet_tag.test.name
							}]
						}
						`, tagName, sourceName, projectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "tags.0.name", tagName),
					resource.TestCheckResourceAttr("sifflet_source.test", "tags.0.kind", "Tag"),
					resource.TestCheckResourceAttrSet("sifflet_source.test", "tags.0.id"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_tag" "test" {
							name = "%s"
						}

						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								bigquery = {
									project_id = "%s"
									dataset_id = "dataset"
									billing_project_id = "dataset"
								}
							}
							tags = [{
								id = sifflet_tag.test.id
							}]
						}
						`, tagName, sourceName, projectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "tags.0.name", tagName),
					resource.TestCheckResourceAttr("sifflet_source.test", "tags.0.kind", "Tag"),
					resource.TestCheckResourceAttrSet("sifflet_source.test", "tags.0.id"),
				),
			},
		},
	})
}

func TestAccSourceParams(t *testing.T) {
	sourceName := randomSourceName()
	project := providertests.RandomName()
	host := providertests.RandomName()
	database := providertests.RandomName()
	catalog := providertests.RandomName()
	accountId := providertests.RandomName()
	clientId := providertests.RandomName()
	accountIdentifier := providertests.RandomName()
	credName := providertests.RandomCredentialName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							parameters = {
								dbt = {
									target = "target"
									project_name = "%s"
								}
							}
						}
						`, sourceName, project),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "dbt"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.dbt.project_name", project),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								airflow = {
									host = "%s"
									port = 3000
								}
							}
						}
						`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "airflow"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.airflow.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							parameters = {
								athena = {
									database = "%s"
									datasource = "datasource"
									region = "region"
									role_arn = "arn:aws:iam::123456789012:role/role"
									s3_output_location = "s3://mybucket"
									workgroup = "workgroup"
								}
							}
						}
						`, sourceName, database),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "athena"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.athena.database", database),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								databricks = {
									catalog = "%s"
									host = "host"
									http_path = "http_path"
									port = 32
									schema = "schema"
								}
							}
						}
						`, sourceName, catalog),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "databricks"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.databricks.catalog", catalog),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								dbt_cloud = {
									account_id = "%s"
									base_url = "base_url"
									project_id = "project_id"
								}
							}
						}
						`, sourceName, accountId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "dbt_cloud"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.dbt_cloud.account_id", accountId),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
							    fivetran	 = { }
							}
						}
						`, sourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "fivetran"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.fivetran.host", "https://api.fivetran.com"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								hive = {
									atlas_base_url = "atlas_base_url"
									atlas_principal = "atlas_principal"
									database = "%s"
									jdbc_url = "jdbc_url"
									krb5_conf = "krb5_conf"
									principal = "principal"
								}
							}
						}
						`, sourceName, database),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "hive"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.hive.database", database),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								looker = {
									host = "%s"
									git_connections = [{
										auth_type = "HTTP_AUTHORIZATION_HEADER"
										branch = "branch"
										secret_id = sifflet_credential.test.name
										url = "url"
									}]
								}
							}
						}
						`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "looker"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.looker.host", host),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.looker.git_connections.#", "1"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.looker.git_connections.0.url", "url"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								mssql = {
									host = "%s"
									database = "database"
									port = 65000
									schema = "schema"
									ssl = false
								}
							}
						}
						`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "mssql"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.mssql.host", host),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.mssql.ssl", "false"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								mysql = {
									host = "%s"
									database = "database"
									port = 65000
									mysql_tls_version = "TLS_V_1_2"
								}
							}
						}
						`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "mysql"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.mysql.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								oracle = {
									host = "%s"
									database = "database"
									port = 65000
									schema = "schema"
								}
							}
						}
						`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "oracle"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.oracle.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								postgresql = {
									host = "%s"
									database = "database"
									port = 65000
									schema = "schema"
								}
							}
						}
						`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "postgresql"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.postgresql.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								power_bi = {
									client_id = "%s"
									tenant_id = "tenant_id"
									workspace_id = "workspace_id"
								}
							}
						}
						`, sourceName, clientId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "power_bi"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.power_bi.client_id", clientId),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							parameters = {
								quicksight = {
									account_id = "%s"
									aws_region = "eu-west-1"
									role_arn = "arn:aws:iam::123456789012:role/role"
								}
							}
						}
						`, sourceName, accountId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "quicksight"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.quicksight.account_id", accountId),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								redshift = {
									host = "%s"
									database = "database"
									port = 65000
									schema = "schema"
								}
							}
						}
						`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "redshift"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.redshift.host", host),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.redshift.ssl", "true"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								snowflake = {
									account_identifier = "%s"
									database = "database"
									schema = "schema"
									warehouse = "warehouse"
								}
							}
						}
						`, sourceName, accountIdentifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "snowflake"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.snowflake.account_identifier", accountIdentifier),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								synapse = {
									host = "%s"
									database = "database"
									schema = "schema"
									port = 65000
								}
							}
						}
						`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "synapse"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.synapse.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								tableau = {
									host = "%s"
								}
							}
						}
						`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "tableau"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.tableau.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credential = sifflet_credential.test.name
							parameters = {
								tableau = {
									host = "%s"
									site = "something"
								}
							}
						}
						`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "tableau"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.tableau.host", host),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.tableau.site", "something"),
				),
			},
		},
	})
}

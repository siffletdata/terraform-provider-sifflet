package source_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider"
	"terraform-provider-sifflet/internal/provider/providertests"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func init() {
	resource.AddTestSweepers("sifflet_source", &resource.Sweeper{
		Name: "sifflet_source",
		F: func(region string) error {
			ctx := context.Background()
			client, err := provider.ClientForSweepers(ctx)
			if err != nil {
				return fmt.Errorf("Error creating HTTP client: %s", err)
			}

			prefix := providertests.AcceptanceTestPrefix()
			filterDto := sifflet.PublicSourceFilterDto{
				TextSearch: &prefix,
			}

			// We only query the first page of results as there are not many dangling resources
			// and we want to keep this sweeper simple.
			var page int32 = 0
			var itemsPerPage int32 = 100

			paginationDto := sifflet.PublicSourcePaginationDto{
				ItemsPerPage: &itemsPerPage,
				Page:         &page,
			}
			requestDto := sifflet.PublicSourceSearchCriteriaDto{
				Filter:     &filterDto,
				Pagination: &paginationDto,
			}

			// Call get sources API a first time to get number of sources to delete
			searchResponse, err := client.PublicGetSourcesWithResponse(ctx, requestDto)
			if err != nil {
				return fmt.Errorf("Error listing sources: %s", err)
			}
			if searchResponse.StatusCode() != http.StatusOK {
				return fmt.Errorf("Error listing sources: status code %d", searchResponse.StatusCode())
			}
			for _, source := range searchResponse.JSON200.Data {
				if strings.HasPrefix(source.Name, providertests.AcceptanceTestPrefix()) {
					_, err = client.PublicDeleteSourceByIdWithResponse(ctx, source.Id)
					if err != nil {
						return fmt.Errorf("Error deleting source %s: %s", source.Name, err)
					}
					fmt.Printf("Deleted dangling source %s\n", source.Name)
				}
			}
			return nil
		},
	})
}

func randomSourceName() string {
	return providertests.RandomName()
}

func baseConfig(credName string) string {
	return providertests.ProviderConfig() + fmt.Sprintf(`
    resource "sifflet_credentials" "test" {
		name = "%s"
		description = "Description"
		value = "Value"
	}
	`, credName)
}

func tagsConfig(tagName string) string {
	return fmt.Sprintf(`
	resource "sifflet_tag" "tag" {
		name = "%s"
		description = "Description"
	}
	`, tagName)
}

// BigQuery sources are also used for testing specific attributes and behaviours of the sifflet_source resource.
// For other source types, we only do a simple create/destroy test.
func TestAccSourceBasic(t *testing.T) {
	sourceName := randomSourceName()
	projectId := providertests.RandomName()
	credName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credentials = sifflet_credentials.test.name
							parameters = {
								bigquery = {
									project_id = "%s"
									dataset_id = "dataset"
									billing_project_id = "dataset"
								}
							}
							timeouts = {
								create = "1m"
								read = "1m"
								update = "1m"
								delete = "1m"
							}
						}
						`, sourceName, projectId),
				// Check that the source_type attribute is known even before the source is created
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue("sifflet_source.test",
							tfjsonpath.New("parameters").AtMapKey("source_type"),
							knownvalue.StringExact("bigquery"),
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("sifflet_source.test", "id"),
					resource.TestCheckResourceAttr("sifflet_source.test", "name", sourceName),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "bigquery"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.bigquery.project_id", projectId),
				),
			},
			// Test description update, should not trigger replacement
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "Another description"
							credentials = sifflet_credentials.test.name
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_source.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Test parameters update, same source
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "Another description"
							credentials = sifflet_credentials.test.name
							parameters = {
								bigquery = {
									project_id = "%s"
									dataset_id = "dataset_2"
									billing_project_id = "dataset"
								}
							}
							timezone = "UTC+1"
						}
						`, sourceName, projectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.bigquery.dataset_id", "dataset_2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_source.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Test parameters update, different source (requires replacement)
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "Another description"
							credentials = sifflet_credentials.test.name
							parameters = {
								snowflake = {
									account_identifier = "%s"
									database = "database"
									schema = "schema"
									warehouse = "warehouse"
								}
							}
							timezone = "UTC+1"
						}
						`, sourceName, projectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.snowflake.database", "database"),
					resource.TestCheckResourceAttr("sifflet_source.test", "parameters.source_type", "snowflake"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_source.test", plancheck.ResourceActionDestroyBeforeCreate),
						// Check that the source_type attribute is known even before the source is created
						plancheck.ExpectKnownValue("sifflet_source.test",
							tfjsonpath.New("parameters").AtMapKey("source_type"),
							knownvalue.StringExact("snowflake"),
						),
					},
				},
			},
		},
	})
}

func TestAccSourceInvalidConfig(t *testing.T) {
	sourceName := randomSourceName()
	projectId := providertests.RandomName()
	database := providertests.RandomName()
	credName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credentials = sifflet_credentials.test.name
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
				ExpectError: regexp.MustCompile("Credentials are required for this source type, but got an empty string"),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							// Credentials, even if they are not required for Athena datasources
							credentials = sifflet_credentials.test.name
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

				ExpectError: regexp.MustCompile("Credentials are not required for this source type and would be ignored"),
			},
		},
	})
}

func TestAccSourceTags(t *testing.T) {
	sourceName := randomSourceName()
	projectId := providertests.RandomName()
	tagName := providertests.RandomName()
	credName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source" "test" {
							name = "%s"
							description = "A description"
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
	credName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
							parameters = {
								looker = {
									host = "%s"
									git_connections = [{
										auth_type = "HTTP_AUTHORIZATION_HEADER"
										branch = "branch"
										secret_id = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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
							credentials = sifflet_credentials.test.name
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

func TestAccMoveFromBigQueryDatasource(t *testing.T) {
	projectId := providertests.RandomName()
	tagName := providertests.RandomName()
	credName := providertests.RandomCredentialsName()
	sourceName := randomSourceName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			// Not possible to move between different resource types before this version
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + tagsConfig(tagName) + fmt.Sprintf(`
					resource "sifflet_datasource" "test" {
						name = "%s"
						secret_id = sifflet_credentials.test.name
						cron_expression = "@daily"
						bigquery = {
							project_id = "%s"
							dataset_id = "dataset"
							billing_project_id = "dataset"
							timezone = {
								timezone = "UTC",
								utc_offset = "(UTC+00:00)'"
							}
						}
						tags = [ sifflet_tag.tag.id ]
					}`, sourceName, projectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("sifflet_datasource.test", "id"),
				),
			},
			{
				Config: baseConfig(credName) + tagsConfig(tagName) + fmt.Sprintf(`
					resource "sifflet_source" "test" {
						name = "%s"
						credentials = sifflet_credentials.test.name
						parameters = {
							bigquery = {
								project_id = "%s"
								dataset_id = "dataset"
								billing_project_id = "dataset"
								timezone = "UTC"
							}
						}
						schedule = "@daily"
						tags = [ {id: sifflet_tag.tag.id} ]
					}

					moved {
						from = sifflet_datasource.test
						to = sifflet_source.test
					}`, sourceName, projectId),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_source.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccMoveFromDbtDatasource(t *testing.T) {
	projectId := providertests.RandomName()
	tagName := providertests.RandomName()
	credName := providertests.RandomCredentialsName()
	sourceName := randomSourceName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			// Not possible to move between different resource types before this version
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + tagsConfig(tagName) + fmt.Sprintf(`
					resource "sifflet_datasource" "test" {
						name = "%s"
						secret_id = sifflet_credentials.test.name
						cron_expression = "@daily"
						dbt = {
							project_name = "%s"
							target = "prod"
						}
						tags = [ sifflet_tag.tag.id ]
					}`, sourceName, projectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("sifflet_datasource.test", "id"),
				),
			},
			{
				Config: baseConfig(credName) + tagsConfig(tagName) + fmt.Sprintf(`
					resource "sifflet_source" "test" {
						name = "%s"
						credentials = sifflet_credentials.test.name
						parameters = {
							dbt = {
								project_name = "%s"
								target = "prod"
							}
						}
						schedule = "@daily"
						tags = [ {id: sifflet_tag.tag.id} ]
					}

					moved {
						from = sifflet_datasource.test
						to = sifflet_source.test
					}`, sourceName, projectId),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_source.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccMoveFromSnowflakeDatasource(t *testing.T) {
	projectId := providertests.RandomName()
	tagName := providertests.RandomName()
	credName := providertests.RandomCredentialsName()
	sourceName := randomSourceName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			// Not possible to move between different resource types before this version
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + tagsConfig(tagName) + fmt.Sprintf(`
					resource "sifflet_datasource" "test" {
						name = "%s"
						secret_id = sifflet_credentials.test.name
						cron_expression = "@daily"
						snowflake = {
							account_identifier = "my-account-id"
							database           = "%s"
							schema             = "schema"
							warehouse          = "warehouse"
						}
						tags = [ sifflet_tag.tag.id ]
					}`, sourceName, projectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("sifflet_datasource.test", "id"),
				),
			},
			{
				Config: baseConfig(credName) + tagsConfig(tagName) + fmt.Sprintf(`
					resource "sifflet_source" "test" {
						name = "%s"
						credentials = sifflet_credentials.test.name
						parameters = {
							snowflake = {
								account_identifier = "my-account-id"
								database           = "%s"
								schema             = "schema"
								warehouse          = "warehouse"
							}
						}
						schedule = "@daily"
						tags = [ {id: sifflet_tag.tag.id} ]
					}

					moved {
						from = sifflet_datasource.test
						to = sifflet_source.test
					}`, sourceName, projectId),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_source.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccSourceTimeout(t *testing.T) {
	sourceName := randomSourceName()
	projectId := providertests.RandomName()
	credName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
					resource "sifflet_source" "timeout_test" {
						name = "%s"
						description = "Testing timeout behavior"
						credentials = sifflet_credentials.test.name
						parameters = {
							bigquery = {
								project_id = "%s"
								dataset_id = "dataset"
								billing_project_id = "dataset"
							}
						}
						timeouts = {
							create = "1ms"  # Extremely short timeout to trigger timeout behavior
							read = "1ms"
							update = "1ms"
							delete = "1ms"
						}
					}
					`, sourceName, projectId),
				ExpectError: regexp.MustCompile("context deadline exceeded|timeout"),
			},
		},
	})
}

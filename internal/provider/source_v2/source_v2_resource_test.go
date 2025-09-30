package source_v2_test

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
)

func init() {
	resource.AddTestSweepers("sifflet_source_v2", &resource.Sweeper{
		Name: "sifflet_source_v2",
		F: func(region string) error {
			ctx := context.Background()
			client, err := provider.ClientForSweepers(ctx)
			if err != nil {
				return fmt.Errorf("Error creating HTTP client: %s", err)
			}

			var page int32 = 0
			var itemsPerPage int32 = 100
			params := sifflet.PublicGetSourcesV2Params{
				Page:         &page,
				ItemsPerPage: &itemsPerPage,
			}

			sources, err := client.PublicGetSourcesV2WithResponse(ctx, &params)
			if err != nil {
				return fmt.Errorf("Error listing sources: %s", err)
			}
			if sources.StatusCode() != 200 {
				return fmt.Errorf("Error listing source: status code %s", sources.Status())
			}
			for _, source := range sources.JSON200.Data {
				var typedSource sifflet.SiffletPublicGetSourceV2Dto
				err := typedSource.FromPublicPageDtoPublicGetSourceV2DtoDataItem(source)
				if err != nil {
					return fmt.Errorf("Error reading source: %s", err)
				}
				sourceDto, err := typedSource.GetSourceDto()
				if err != nil {
					return fmt.Errorf("Error reading source: %s", err)
				}
				if strings.HasPrefix(sourceDto.GetName(), providertests.AcceptanceTestPrefix()) {
					source_id := sourceDto.GetId()
					deleteResponse, err := client.PublicDeleteSourceV2WithResponse(ctx, source_id)
					if err != nil {
						return fmt.Errorf("Error deleting source %s: %s", sourceDto.GetName(), err)
					}
					if deleteResponse.StatusCode() != http.StatusNoContent {
						return fmt.Errorf("Error deleting source: %s status code %d", deleteResponse.Status(), deleteResponse.StatusCode())
					}
					fmt.Printf("Deleted dangling source v2 with id %s and name %s\n", source_id, sourceDto.GetName())
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

// MySQL sources are also used for testing specific attributes and behaviours of the sifflet_source resource.
// For other source types, we only do a simple create/destroy test.
func TestAccSourceV2Basic(t *testing.T) {
	sourceName := randomSourceName()
	hostId := providertests.RandomName()
	credName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source_v2" "test" {
							name = "%s-1"
							parameters = {
								mysql = {
									host = "%s"
									port = "3306"
									database = "database"
									mysql_tls_version = "TLS_V_1_2"
									credentials = sifflet_credentials.test.name
								}
							}
						}
						`, sourceName, hostId),
				// Check that the source_type attribute is known even before the source is created
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue("sifflet_source_v2.test",
							tfjsonpath.New("parameters").AtMapKey("source_type"),
							knownvalue.StringExact("mysql"),
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("sifflet_source_v2.test", "id"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-1", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "mysql"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.mysql.host", hostId),
				),
			},
			// Test database name and timezone update, should not trigger replacement
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source_v2" "test" {
							name = "%s-2"
							parameters = {
								mysql = {
									host = "%s"
									port = "3306"
									database = "database_new"
									mysql_tls_version = "TLS_V_1_2"
									credentials = sifflet_credentials.test.name
								}
							}
							timezone = "GMT"
						}
						`, sourceName, hostId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.mysql.database", "database_new"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "timezone", "GMT"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_source_v2.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			// Test parameters update, different source (requires replacement)
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source_v2" "test" {
							name = "%s-3"
							parameters = {
								tableau = {
									site = "%s"
									host = "host"
									credentials = sifflet_credentials.test.name
								}
							}
							timezone = "GMT"
						}
						`, sourceName, hostId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.tableau.host", "host"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "tableau"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("sifflet_source_v2.test", plancheck.ResourceActionDestroyBeforeCreate),
						// Check that the source_type attribute is known even before the source is created
						plancheck.ExpectKnownValue("sifflet_source_v2.test",
							tfjsonpath.New("parameters").AtMapKey("source_type"),
							knownvalue.StringExact("tableau"),
						),
					},
				},
			},
		},
	})
}

// Test all data source types
func TestAccSourceInvalidConfigV2(t *testing.T) {
	sourceName := randomSourceName()
	hostId := providertests.RandomName()
	credName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
						resource "sifflet_source_v2" "test" {
							name = "%s"
							parameters = {
								mysql = {
									host = "%s"
									port = "3306"
									database = "database"
									mysql_tls_version = "TLS_V_1_2"
									credentials = sifflet_credentials.test.name
								}
								databricks = {
									host = "host"
									http_path = "http_path"
									port = 32
									credentials = sifflet_credentials.test.name
								}
							}
						}
						`, sourceName, hostId),
				ExpectError: regexp.MustCompile("Exactly one of these attributes must be configured"),
			},
		},
	})
}

func TestAccSourceParamsV2(t *testing.T) {
	sourceName := randomSourceName()
	project := providertests.RandomName()
	host := providertests.RandomName()
	accountId := providertests.RandomName()
	clientId := providertests.RandomName()
	accountIdentifier := providertests.RandomName()
	credName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-dbt"
				parameters = {
					dbt = {
						target = "target"
						project_name = "%s"
					}
				}
			}
			`, sourceName, project),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-dbt", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "dbt"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.dbt.project_name", project),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-airflow"
				parameters = {
					airflow = {
						host = "%s"
						port = 3000
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-airflow", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "airflow"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.airflow.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-athena"
				parameters = {
					athena = {
						datasource = "datasource"
						region = "region"
						role_arn = "arn:aws:iam::123456789012:role/role"
						s3_output_location = "s3://mybucket"
						workgroup = "workgroup"
					}
				}
			}
			`, sourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-athena", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "athena"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.athena.datasource", "datasource"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-databricks"
				parameters = {
					databricks = {
						host = "%s"
						http_path = "http_path"
						port = 32
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-databricks", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "databricks"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.databricks.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-dbt-cloud"
				parameters = {
					dbtcloud = {
						account_id = "%s"
						base_url = "base_url"
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, accountId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-dbt-cloud", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "dbtcloud"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.dbtcloud.account_id", accountId),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-fivetran"
				parameters = {
				    fivetran	 = {
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-fivetran", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "fivetran"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.fivetran.host", "https://api.fivetran.com"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-looker-1"
				parameters = {
					looker = {
					    credentials = sifflet_credentials.test.name
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
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-looker-1", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "looker"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.looker.host", host),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.looker.git_connections.#", "1"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.looker.git_connections.0.url", "url"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-looker-2"
				parameters = {
					looker = {
					    credentials = sifflet_credentials.test.name
						host = "%s"
						git_connections = [{
							auth_type = "HTTP_AUTHORIZATION_HEADER"
							secret_id = sifflet_credentials.test.name
							url = "url"
						}]
					}
				}
			}
			`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-looker-2", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "looker"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.looker.host", host),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.looker.git_connections.#", "1"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.looker.git_connections.0.url", "url"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-mssql"
				parameters = {
					mssql = {
						host = "%s"
						database = "database"
						port = 65000
						ssl = false
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-mssql", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "mssql"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.mssql.host", host),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.mssql.ssl", "false"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-mysql"
				parameters = {
					mysql = {
						host = "%s"
						database = "database"
						port = 65000
						mysql_tls_version = "TLS_V_1_2"
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-mysql", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "mysql"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.mysql.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-oracle"
				parameters = {
					oracle = {
						host = "%s"
						database = "database"
						port = 65000
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-oracle", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "oracle"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.oracle.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-postgresql"
				parameters = {
					postgresql = {
						host = "%s"
						database = "database"
						port = 65000
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-postgresql", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "postgresql"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.postgresql.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-power-bi"
				parameters = {
					power_bi = {
						client_id = "%s"
						tenant_id = "tenant_id"
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, clientId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-power-bi", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "power_bi"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.power_bi.client_id", clientId),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-quicksight"
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
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-quicksight", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "quicksight"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.quicksight.account_id", accountId),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-redshift"
				parameters = {
					redshift = {
						host = "%s"
						port = 65000
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-redshift", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "redshift"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.redshift.host", host),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.redshift.ssl", "true"),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-snowflake"
				parameters = {
					snowflake = {
						account_identifier = "%s"
						warehouse = "warehouse"
				        credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, accountIdentifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-snowflake", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "snowflake"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.snowflake.account_identifier", accountIdentifier),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-synapse"
				parameters = {
					synapse = {
						host = "%s"
						port = 65000
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-synapse", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "synapse"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.synapse.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-tableau-1"
				parameters = {
					tableau = {
						host = "%s"
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-tableau-1", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "tableau"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.tableau.host", host),
				),
			},
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
			resource "sifflet_source_v2" "test" {
				name = "%s-tableau-2"
				parameters = {
					tableau = {
						host = "%s"
						site = "something"
						credentials = sifflet_credentials.test.name
					}
				}
			}
			`, sourceName, host),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "name", fmt.Sprintf("%s-tableau-2", sourceName)),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.source_type", "tableau"),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.tableau.host", host),
					resource.TestCheckResourceAttr("sifflet_source_v2.test", "parameters.tableau.site", "something"),
				),
			},
		},
	})
}

func TestAccSourceV2Timeout(t *testing.T) {
	sourceName := randomSourceName()
	host := providertests.RandomName()
	credName := providertests.RandomCredentialsName()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: provider.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: baseConfig(credName) + fmt.Sprintf(`
					resource "sifflet_source_v2" "test" {
						name = "%s"
						parameters = {
							mysql = {
								host = "%s"
								port = "3306"
								database = "database"
								mysql_tls_version = "TLS_V_1_2"
								credentials = sifflet_credentials.test.name
							}
						}
						timeouts = {
							create = "1ms"  # Extremely short timeout to trigger timeout behavior
							read = "1ms"
							update = "1ms"
							delete = "1ms"
						}
					}
					`, sourceName, host),
				ExpectError: regexp.MustCompile("context deadline exceeded|timeout"),
			},
		},
	})
}

---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "sifflet_sources Data Source - terraform-provider-sifflet"
subcategory: ""
description: |-
  Return sources matching search criteria.
---

# sifflet_sources (Data Source)

Return sources matching search criteria.

## Example Usage

```terraform
data "sifflet_sources" "test" {
  filter = {
    text_search = "source_name"
    types       = ["MYSQL"]
    tags = [{
      name = "tag_name"
    }]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `filter` (Attributes) Search criteria (see [below for nested schema](#nestedatt--filter))

### Optional

- `max_results` (Number) Maximum number of results to return. Default is 1000.

### Read-Only

- `results` (Attributes List) List of sources returned by the search (see [below for nested schema](#nestedatt--results))

<a id="nestedatt--filter"></a>
### Nested Schema for `filter`

Optional:

- `tags` (Attributes List) List of tags to filter sources by. Tags can be identified by either ID or name. If a name is provided, optionally a kind can be provided to disambiguate tags of different types sharing the same name. (see [below for nested schema](#nestedatt--filter--tags))
- `text_search` (String) Return sources whose name match this attribute
- `types` (List of String) List of source types to filter sources by

<a id="nestedatt--filter--tags"></a>
### Nested Schema for `filter.tags`

Optional:

- `id` (String) Tag ID
- `kind` (String) Tag kind (such as 'Tag' or 'Classification')
- `name` (String) Tag name



<a id="nestedatt--results"></a>
### Nested Schema for `results`

Read-Only:

- `credentials` (String) Name of the credentials used to connect to the source
- `description` (String) Source description
- `id` (String) Source ID
- `name` (String) Source name
- `parameters` (Attributes) Connection parameters. Provide only one nested block depending on the source type. (see [below for nested schema](#nestedatt--results--parameters))
- `schedule` (String) When this source is scheduled to refresh (cron expression). Will be null if not scheduled.
- `tags` (Attributes List) List of tags associated with this source (see [below for nested schema](#nestedatt--results--tags))
- `timezone` (String) Timezone for this source

<a id="nestedatt--results--parameters"></a>
### Nested Schema for `results.parameters`

Optional:

- `airflow` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--airflow))
- `athena` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--athena))
- `bigquery` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--bigquery))
- `databricks` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--databricks))
- `dbt` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--dbt))
- `dbt_cloud` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--dbt_cloud))
- `fivetran` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--fivetran))
- `looker` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--looker))
- `mssql` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--mssql))
- `mysql` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--mysql))
- `oracle` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--oracle))
- `postgresql` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--postgresql))
- `power_bi` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--power_bi))
- `quicksight` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--quicksight))
- `redshift` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--redshift))
- `snowflake` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--snowflake))
- `synapse` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--synapse))
- `tableau` (Attributes) (see [below for nested schema](#nestedatt--results--parameters--tableau))

Read-Only:

- `source_type` (String) Source type (e.g bigquery, dbt, ...). This attribute is automatically set depending on which connection parameters are set.

<a id="nestedatt--results--parameters--airflow"></a>
### Nested Schema for `results.parameters.airflow`

Required:

- `host` (String) Airflow API host
- `port` (Number) Airflow API port


<a id="nestedatt--results--parameters--athena"></a>
### Nested Schema for `results.parameters.athena`

Required:

- `database` (String) Athena database name
- `datasource` (String) Athena datasource name
- `region` (String) AWS region in which the Athena database is located
- `role_arn` (String) AWS IAM role ARN to use for Athena queries
- `s3_output_location` (String) S3 location to store Athena query results
- `workgroup` (String) Athena workgroup name

Optional:

- `vpc_url` (String) VPC URL for Athena queries


<a id="nestedatt--results--parameters--bigquery"></a>
### Nested Schema for `results.parameters.bigquery`

Required:

- `dataset_id` (String) BigQuery dataset ID
- `project_id` (String) GCP project ID containing the BigQuery dataset.

Optional:

- `billing_project_id` (String) GCP billing project ID


<a id="nestedatt--results--parameters--databricks"></a>
### Nested Schema for `results.parameters.databricks`

Required:

- `catalog` (String) Databricks catalog name
- `host` (String) Databricks host
- `http_path` (String) Databricks HTTP path
- `port` (Number) Databricks server port
- `schema` (String) Databricks schema


<a id="nestedatt--results--parameters--dbt"></a>
### Nested Schema for `results.parameters.dbt`

Required:

- `project_name` (String) dbt project name
- `target` (String) dbt target name (the 'target' value in the profiles.yml file)


<a id="nestedatt--results--parameters--dbt_cloud"></a>
### Nested Schema for `results.parameters.dbt_cloud`

Required:

- `account_id` (String) dbt Cloud account ID
- `base_url` (String) dbt Cloud base URL
- `project_id` (String) dbt Cloud project ID

Optional:

- `job_definition_id` (String) dbt Cloud job definition ID


<a id="nestedatt--results--parameters--fivetran"></a>
### Nested Schema for `results.parameters.fivetran`

Optional:

- `host` (String) Fivetran host. Defaults to https://api.fivetran.com.


<a id="nestedatt--results--parameters--looker"></a>
### Nested Schema for `results.parameters.looker`

Required:

- `git_connections` (Attributes List) Configuration for the repositories storing LookML code. See [the Sifflet documentation](https://docs.siffletdata.com/docs/looker) for details. If you don't use LookML, pass an empty list. (see [below for nested schema](#nestedatt--results--parameters--looker--git_connections))
- `host` (String) URL of the Looker API for your instance. If your Looker instance is hosted at https://mycompany.looker.com, the API URL is https://mycompany.looker.com/api/4.0

<a id="nestedatt--results--parameters--looker--git_connections"></a>
### Nested Schema for `results.parameters.looker.git_connections`

Required:

- `auth_type` (String) Authentication type for the Git connection. Valid values are 'HTTP_AUTHORIZATION_HEADER', 'USER_PASSWORD' or 'SSH'. See the Sifflet docs for the meaning of each value.
- `secret_id` (String) Secret (credential) ID to use for authentication. The secret contents must match the chosen authentication type: access token for 'HTTP_AUTHORIZATION_HEADER' or 'USER_PASSWORD', or private SSH key for 'SSH'. See the Sifflet docs for more details.
- `url` (String) URL of the Git repository containing the LookML code.

Optional:

- `branch` (String) Branch of the Git repository to use. If omitted, the default branch is used.



<a id="nestedatt--results--parameters--mssql"></a>
### Nested Schema for `results.parameters.mssql`

Required:

- `database` (String) Database name
- `host` (String) Microsoft SQL Server hostname
- `port` (Number) Microsoft SQL Server port number
- `schema` (String) Schema name

Optional:

- `ssl` (Boolean, Deprecated) Use TLS to connect to Microsoft SQL Server.


<a id="nestedatt--results--parameters--mysql"></a>
### Nested Schema for `results.parameters.mysql`

Required:

- `database` (String) Database name
- `host` (String) MySQL server hostname
- `mysql_tls_version` (String) TLS version to use for MySQL connection. One of TLS_V_1_2 or TLS_V_1_3.
- `port` (Number) MySQL port number


<a id="nestedatt--results--parameters--oracle"></a>
### Nested Schema for `results.parameters.oracle`

Required:

- `database` (String) Database name
- `host` (String) Oracle server hostname
- `port` (Number) Oracle server port number
- `schema` (String) Schema name


<a id="nestedatt--results--parameters--postgresql"></a>
### Nested Schema for `results.parameters.postgresql`

Required:

- `database` (String) Database name
- `host` (String) PostgreSQL server hostname
- `port` (Number) PostgreSQL server port number
- `schema` (String) Schema name


<a id="nestedatt--results--parameters--power_bi"></a>
### Nested Schema for `results.parameters.power_bi`

Required:

- `client_id` (String) Azure AD client ID
- `tenant_id` (String) Azure AD tenant ID
- `workspace_id` (String) Power BI workspace ID


<a id="nestedatt--results--parameters--quicksight"></a>
### Nested Schema for `results.parameters.quicksight`

Required:

- `account_id` (String) AWS account ID
- `aws_region` (String) AWS region
- `role_arn` (String) AWS IAM role ARN used to access QuickSight (see Sifflet documentation for details)


<a id="nestedatt--results--parameters--redshift"></a>
### Nested Schema for `results.parameters.redshift`

Required:

- `database` (String) Database name
- `host` (String) Redshift server hostname
- `port` (Number) Redshift server port number
- `schema` (String) Schema name

Optional:

- `ssl` (Boolean) Use TLS to connect to Redshift. It's strongly recommended to keep this option enabled.


<a id="nestedatt--results--parameters--snowflake"></a>
### Nested Schema for `results.parameters.snowflake`

Required:

- `account_identifier` (String) Snowflake account identifier
- `database` (String) Database name
- `schema` (String) Schema name
- `warehouse` (String) Warehouse name, used by Sifflet to run queries


<a id="nestedatt--results--parameters--synapse"></a>
### Nested Schema for `results.parameters.synapse`

Required:

- `database` (String) Database name
- `host` (String) Azure Synapse server hostname
- `port` (Number) Azure Synapse server port number
- `schema` (String) Schema name


<a id="nestedatt--results--parameters--tableau"></a>
### Nested Schema for `results.parameters.tableau`

Required:

- `host` (String) Tableau Server hostname

Optional:

- `site` (String) Tableau Server site. If your Tableau environment is using the Default site, omit this field.



<a id="nestedatt--results--tags"></a>
### Nested Schema for `results.tags`

Read-Only:

- `id` (String) Tag ID
- `kind` (String) Tag kind (such as 'Tag' or 'Classification')
- `name` (String) Tag name

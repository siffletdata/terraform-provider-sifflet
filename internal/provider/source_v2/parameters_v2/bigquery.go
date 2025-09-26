package parameters_v2

import (
	"context"
	"encoding/json"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/provider/source_v2/parameters_v2/scope"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BigQueryParametersModel struct {
	ProjectId        types.String `tfsdk:"project_id"`
	BillingProjectId types.String `tfsdk:"billing_project_id"`
	Credentials      types.String `tfsdk:"credentials"`
	Schedule         types.String `tfsdk:"schedule"`
	Scope            types.Object `tfsdk:"scope"`
}

func (m BigQueryParametersModel) SchemaSourceType() string {
	return "bigquery"
}

func (m BigQueryParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "GCP project ID containing the BigQuery dataset.",
				Required:    true,
			},
			"billing_project_id": schema.StringAttribute{
				Description: "GCP billing project ID",
				Optional:    true,
			},
			"credentials": schema.StringAttribute{
				Description: "Name of the credentials used to connect to the source.",
				Required:    true,
			},
			"schedule": schema.StringAttribute{
				Description: "Schedule for the source. Must be a valid cron expression. If empty, the source will only be refreshed when manually triggered.",
				Optional:    true,
			},
			"scope": schema.SingleNestedAttribute{
				Description: "Datasets to include or exclude. If not specified, all the datasets will be included (including future ones).",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Whether to include or exclude the specified datasets. One of INCLUSION or EXCLUSION.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("INCLUSION", "EXCLUSION"),
						},
					},
					"datasets": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "The datasets to either include or exclude.",
					},
				},
			},
		},
	}
}

func (m BigQueryParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"project_id":         types.StringType,
		"billing_project_id": types.StringType,
		"credentials":        types.StringType,
		"schedule":           types.StringType,
		"scope":              scope.DatasetsScopeTypeAttributes,
	}
}

func (m BigQueryParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	bigqueryParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.BigQuery = bigqueryParams
	return o, diag.Diagnostics{}
}

func (m BigQueryParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	bigQueryInformation := sifflet.BigQueryInformation{
		BillingProjectId: m.BillingProjectId.ValueStringPointer(),
		ProjectId:        m.ProjectId.ValueString(),
	}

	scopeDto, diags := scope.ToPublicDatasetsScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicCreateSourceV2JSONBody{}, diags
	}

	bigQueryCreateDto := &sifflet.PublicCreateBigQuerySourceV2Dto{
		Name:                name,
		Timezone:            &timezone,
		Type:                sifflet.PublicCreateBigQuerySourceV2DtoTypeBIGQUERY,
		BigQueryInformation: &bigQueryInformation,
		Credentials:         m.Credentials.ValueStringPointer(),
		Schedule:            m.Schedule.ValueStringPointer(),
		Scope:               scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(bigQueryCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create BigQuery source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m BigQueryParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	bigQueryInformation := sifflet.BigQueryInformation{
		BillingProjectId: m.BillingProjectId.ValueStringPointer(),
		ProjectId:        m.ProjectId.ValueString(),
	}

	scopeDto, diags := scope.ToPublicDatasetsScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicEditSourceV2JSONBody{}, diags
	}

	bigQueryUpdateDto := &sifflet.PublicUpdateBigQuerySourceV2Dto{
		Name:                &name,
		Timezone:            &timezone,
		Type:                sifflet.PublicUpdateBigQuerySourceV2DtoTypeBIGQUERY,
		BigQueryInformation: bigQueryInformation,
		Credentials:         m.Credentials.ValueString(),
		Schedule:            m.Schedule.ValueStringPointer(),
		Scope:               scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(bigQueryUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update BigQuery source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *BigQueryParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	bigQueryDto := d.PublicGetBigQuerySourceV2Dto
	if bigQueryDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read BigQuery source", "Source does not contain BigQuery params but was interpreted as a BigQuery source")}
	}

	m.ProjectId = types.StringValue(bigQueryDto.BigQueryInformation.ProjectId)
	m.BillingProjectId = types.StringValue(*bigQueryDto.BigQueryInformation.BillingProjectId)
	m.Credentials = types.StringPointerValue(bigQueryDto.Credentials)
	m.Schedule = types.StringPointerValue(bigQueryDto.Schedule)
	scopeObject, diags := scope.FromPublicDatasetsScopeDto(ctx, bigQueryDto.Scope)
	if diags.HasError() {
		return diags
	}
	m.Scope = scopeObject
	return diag.Diagnostics{}
}

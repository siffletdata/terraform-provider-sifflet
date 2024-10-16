package parameters

import (
	"context"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type BigQueryParametersModel struct {
	ProjectId        types.String `tfsdk:"project_id"`
	BillingProjectId types.String `tfsdk:"billing_project_id"`
	DatasetId        types.String `tfsdk:"dataset_id"`
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
			"dataset_id": schema.StringAttribute{
				Description: "BigQuery dataset ID",
				Required:    true,
			},
		},
	}
}

func (m BigQueryParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"project_id":         types.StringType,
		"billing_project_id": types.StringType,
		"dataset_id":         types.StringType,
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

func (m BigQueryParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.BigQuery.IsNull()
}

func (m *BigQueryParametersModel) CreateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.BigQuery.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicBigQueryParametersDto{
		Type:             sifflet.PublicBigQueryParametersDtoTypeBIGQUERY,
		ProjectId:        m.ProjectId.ValueStringPointer(),
		BillingProjectId: m.BillingProjectId.ValueStringPointer(),
		DatasetId:        m.DatasetId.ValueStringPointer(),
	}
	err := parametersDto.FromPublicBigQueryParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *BigQueryParametersModel) UpdateSourceDtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicUpdateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicUpdateSourceDto_Parameters
	diags := p.BigQuery.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diags
	}
	dto := sifflet.PublicBigQueryParametersDto{
		Type:             sifflet.PublicBigQueryParametersDtoTypeBIGQUERY,
		ProjectId:        m.ProjectId.ValueStringPointer(),
		BillingProjectId: m.BillingProjectId.ValueStringPointer(),
		DatasetId:        m.DatasetId.ValueStringPointer(),
	}
	err := parametersDto.FromPublicBigQueryParametersDto(dto)
	if err != nil {
		return sifflet.PublicUpdateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to update source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *BigQueryParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicBigQueryParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.ProjectId = types.StringPointerValue(paramsDto.ProjectId)
	m.BillingProjectId = types.StringPointerValue(paramsDto.BillingProjectId)
	m.DatasetId = types.StringPointerValue(paramsDto.DatasetId)
	return diag.Diagnostics{}
}

func (m BigQueryParametersModel) RequiresCredential() bool {
	return true
}

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

type PowerBiParametersModel struct {
	ClientId    types.String `tfsdk:"client_id"`
	TenantId    types.String `tfsdk:"tenant_id"`
	Credentials types.String `tfsdk:"credentials"`
	Scope       types.Object `tfsdk:"scope"`
	Schedule    types.String `tfsdk:"schedule"`
}

func (m PowerBiParametersModel) SchemaSourceType() string {
	return "power_bi"
}

func (m PowerBiParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Description: "Your Azure AD client ID",
				Required:    true,
			},
			"tenant_id": schema.StringAttribute{
				Description: "Your Azure AD tenant ID",
				Required:    true,
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
				Description: "Workspaces to include or exclude. If not specified, all the workspaces will be included (including future ones).",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Whether to include or exclude the specified workspaces. One of INCLUSION or EXCLUSION.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("INCLUSION", "EXCLUSION"),
						},
					},
					"workspaces": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "The workspaces to either include or exclude.",
					},
				},
			},
		},
	}
}

func (m PowerBiParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"client_id":   types.StringType,
		"tenant_id":   types.StringType,
		"credentials": types.StringType,
		"schedule":    types.StringType,
		"scope":       scope.WorkspacesScopeTypeAttributes,
	}
}

func (m PowerBiParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	powerBiParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.PowerBi = powerBiParams
	return o, diag.Diagnostics{}
}

func (m PowerBiParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	powerBiInformation := sifflet.PowerBiInformation{
		ClientId: m.ClientId.ValueString(),
		TenantId: m.TenantId.ValueString(),
	}

	scopeDto, diags := scope.ToPublicWorkspacesScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicCreateSourceV2JSONBody{}, diags
	}

	powerBiCreateDto := &sifflet.PublicCreatePowerBiSourceV2Dto{
		Name:               name,
		Timezone:           &timezone,
		Type:               sifflet.PublicCreatePowerBiSourceV2DtoTypePOWERBI,
		PowerBiInformation: &powerBiInformation,
		Credentials:        m.Credentials.ValueStringPointer(),
		Schedule:           m.Schedule.ValueStringPointer(),
		Scope:              scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(powerBiCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Power BI source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m PowerBiParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	powerBiInformation := sifflet.PowerBiInformation{
		ClientId: m.ClientId.ValueString(),
		TenantId: m.TenantId.ValueString(),
	}

	scopeDto, diags := scope.ToPublicWorkspacesScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicEditSourceV2JSONBody{}, diags
	}

	powerBiUpdateDto := &sifflet.PublicUpdatePowerBiSourceV2Dto{
		Name:               &name,
		Timezone:           &timezone,
		Type:               sifflet.PublicUpdatePowerBiSourceV2DtoTypePOWERBI,
		PowerBiInformation: powerBiInformation,
		Credentials:        m.Credentials.ValueString(),
		Schedule:           m.Schedule.ValueStringPointer(),
		Scope:              scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(powerBiUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Power BI source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *PowerBiParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	powerBiDto := d.PublicGetPowerBiSourceV2Dto
	if powerBiDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Power BI source", "Source does not contain Power BI params but was interpreted as a Power BI source")}
	}

	m.ClientId = types.StringValue(powerBiDto.PowerBiInformation.ClientId)
	m.TenantId = types.StringValue(powerBiDto.PowerBiInformation.TenantId)
	m.Credentials = types.StringPointerValue(powerBiDto.Credentials)
	m.Schedule = types.StringPointerValue(powerBiDto.Schedule)
	scopeObject, diags := scope.FromPublicWorkspacesScopeDto(ctx, powerBiDto.Scope)
	if diags.HasError() {
		return diags
	}
	m.Scope = scopeObject
	return diag.Diagnostics{}
}

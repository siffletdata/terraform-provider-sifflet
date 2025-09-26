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

type TableauParametersModel struct {
	Host        types.String `tfsdk:"host"`
	Site        types.String `tfsdk:"site"`
	Credentials types.String `tfsdk:"credentials"`
	Schedule    types.String `tfsdk:"schedule"`
	Scope       types.Object `tfsdk:"scope"`
}

func (m TableauParametersModel) SchemaSourceType() string {
	return "tableau"
}

func (m TableauParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Your Tableau Server hostname",
				Required:    true,
			},
			"site": schema.StringAttribute{
				Description: "Your Tableau Server site. Leave empty if your Tableau environment is using the Default Site.",
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
				Description: "Folders to include or exclude. If not specified, all the folders will be included (including future ones).",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Whether to include or exclude the specified folders. One of INCLUSION or EXCLUSION.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("INCLUSION", "EXCLUSION"),
						},
					},
					"folders": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
						Description: "The folders to either include or exclude.",
					},
				},
			},
		},
	}
}

func (m TableauParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":        types.StringType,
		"site":        types.StringType,
		"credentials": types.StringType,
		"schedule":    types.StringType,
		"scope":       scope.FoldersScopeTypeAttributes,
	}
}

func (m TableauParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	tableauParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Tableau = tableauParams
	return o, diag.Diagnostics{}
}

func (m TableauParametersModel) ToCreateDto(ctx context.Context, name string, timezone string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	tableauInformation := sifflet.TableauInformation{
		Host: m.Host.ValueString(),
		Site: m.Site.ValueString(),
	}

	scopeDto, diags := scope.ToPublicFoldersScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicCreateSourceV2JSONBody{}, diags
	}

	tableauCreateDto := &sifflet.PublicCreateTableauSourceV2Dto{
		Name:               name,
		Timezone:           &timezone,
		Type:               sifflet.PublicCreateTableauSourceV2DtoTypeTABLEAU,
		TableauInformation: &tableauInformation,
		Credentials:        m.Credentials.ValueStringPointer(),
		Schedule:           m.Schedule.ValueStringPointer(),
		Scope:              scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(tableauCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Tableau source", err)
	}
	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	createSourceJsonBody.SetRawMessage(buf)

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m TableauParametersModel) ToUpdateDto(ctx context.Context, name string, timezone string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	tableauInformation := sifflet.TableauInformation{
		Host: m.Host.ValueString(),
		Site: m.Site.ValueString(),
	}

	scopeDto, diags := scope.ToPublicFoldersScopeDto(ctx, m.Scope)
	if diags.HasError() {
		return sifflet.PublicEditSourceV2JSONBody{}, diags
	}

	tableauUpdateDto := &sifflet.PublicUpdateTableauSourceV2Dto{
		Name:               &name,
		Timezone:           &timezone,
		Type:               sifflet.PublicUpdateTableauSourceV2DtoTypeTABLEAU,
		TableauInformation: tableauInformation,
		Credentials:        m.Credentials.ValueString(),
		Schedule:           m.Schedule.ValueStringPointer(),
		Scope:              scopeDto,
	}

	// We marshal the DTO to JSON manually since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(tableauUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Tableau source", err)
	}
	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	editSourceJsonBody.SetRawMessage(buf)

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *TableauParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	tableauDto := d.PublicGetTableauSourceV2Dto
	if tableauDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Tableau source", "Source does not contain Tableau params but was interpreted as a Tableau source")}
	}

	m.Host = types.StringValue(tableauDto.TableauInformation.Host)
	m.Site = types.StringValue(tableauDto.TableauInformation.Site)
	m.Credentials = types.StringPointerValue(tableauDto.Credentials)
	m.Schedule = types.StringPointerValue(tableauDto.Schedule)
	scopeObject, diags := scope.FromPublicFoldersScopeDto(ctx, tableauDto.Scope)
	if diags.HasError() {
		return diags
	}
	m.Scope = scopeObject
	return diag.Diagnostics{}
}

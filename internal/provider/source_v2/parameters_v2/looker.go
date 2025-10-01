package parameters_v2

import (
	"context"
	"fmt"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	gitConnectionTypeAttributes = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"auth_type": types.StringType,
			"branch":    types.StringType,
			"secret_id": types.StringType,
			"url":       types.StringType,
		},
	}
)

type gitConnectionModel struct {
	AuthType types.String `tfsdk:"auth_type"`
	Branch   types.String `tfsdk:"branch"`
	SecretId types.String `tfsdk:"secret_id"`
	Url      types.String `tfsdk:"url"`
}

type LookerParametersModel struct {
	GitConnections types.List   `tfsdk:"git_connections"`
	Host           types.String `tfsdk:"host"`
	Credentials    types.String `tfsdk:"credentials"`
	Schedule       types.String `tfsdk:"schedule"`
}

func (m LookerParametersModel) SchemaSourceType() string {
	return "looker"
}

func (m LookerParametersModel) TerraformSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true,
		Attributes: map[string]schema.Attribute{
			"git_connections": schema.ListNestedAttribute{
				Description:         "Configuration for the repositories storing LookML code. See https://docs.siffletdata.com/docs/looker for details. If you don't use LookML, pass an empty list.",
				MarkdownDescription: "Configuration for the repositories storing LookML code. See [the Sifflet documentation](https://docs.siffletdata.com/docs/looker) for details. If you don't use LookML, pass an empty list.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"auth_type": schema.StringAttribute{
							Description: "Authentication type for the Git connection. Valid values are 'HTTP_AUTHORIZATION_HEADER', 'USER_PASSWORD' or 'SSH'. See the Sifflet docs for the meaning of each value.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("HTTP_AUTHORIZATION_HEADER", "USER_PASSWORD", "SSH"),
							},
						},
						"branch": schema.StringAttribute{
							Description: "Branch of the Git repository to use. If omitted, the default branch is used.",
							Optional:    true,
							Computed:    true,
							// If we don't send a branch, the API will set it to an empty string. To be consistent with the plan, we change nil values to empty strings.
							Default: stringdefault.StaticString(""),
						},
						"secret_id": schema.StringAttribute{
							Description: "Secret (credential) ID to use for authentication. The secret contents must match the chosen authentication type: access token for 'HTTP_AUTHORIZATION_HEADER' or 'USER_PASSWORD', or private SSH key for 'SSH'. See the Sifflet docs for more details.",
							Required:    true,
						},
						"url": schema.StringAttribute{
							Description: "URL of the Git repository containing the LookML code.",
							Required:    true,
						},
					},
				},
			},
			"host": schema.StringAttribute{
				Description: "URL of the Looker API for your instance. If your Looker instance is hosted at https://mycompany.looker.com, the API URL is https://mycompany.looker.com/api/4.0",
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
		},
	}
}

func (m LookerParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"git_connections": types.ListType{ElemType: gitConnectionTypeAttributes},
		"host":            types.StringType,
		"credentials":     types.StringType,
		"schedule":        types.StringType,
	}
}

func (m LookerParametersModel) AsParametersModel(ctx context.Context) (ParametersModel, diag.Diagnostics) {
	lookerParams, diags := types.ObjectValueFrom(ctx, m.AttributeTypes(), m)
	if diags.HasError() {
		return ParametersModel{}, diags
	}
	o := NewParametersModel()
	o.Looker = lookerParams
	return o, diag.Diagnostics{}
}

func parseGitConnectionAuthType(s string) (sifflet.GitConnectionAuthType, error) {
	switch s {
	case "HTTP_AUTHORIZATION_HEADER":
		return sifflet.HTTPAUTHORIZATIONHEADER, nil
	case "USER_PASSWORD":
		return sifflet.USERPASSWORD, nil
	case "SSH":
		return sifflet.SSH, nil
	default:
		return sifflet.GitConnectionAuthType(""), fmt.Errorf("invalid Git connection auth type: %s", s)
	}
}

func gitConnectionAuthTypeToString(t sifflet.GitConnectionAuthType) (string, error) {
	switch t {
	case sifflet.HTTPAUTHORIZATIONHEADER:
		return "HTTP_AUTHORIZATION_HEADER", nil
	case sifflet.USERPASSWORD:
		return "USER_PASSWORD", nil
	case sifflet.SSH:
		return "SSH", nil
	default:
		return "", fmt.Errorf("invalid Git connection auth type: %s", t)
	}
}

func makeGitConnectionsDto(ctx context.Context, m LookerParametersModel) ([]sifflet.GitConnection, diag.Diagnostics) {
	gitConnectionObjects := make([]types.Object, 0, len(m.GitConnections.Elements()))
	diags := m.GitConnections.ElementsAs(ctx, &gitConnectionObjects, false)
	if diags.HasError() {
		return []sifflet.GitConnection{}, diags
	}

	gitConnections := make([]sifflet.GitConnection, len(gitConnectionObjects))
	for i, gitConnectionObj := range gitConnectionObjects {
		var gitConnectionModel gitConnectionModel
		diags := gitConnectionObj.As(ctx, &gitConnectionModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return []sifflet.GitConnection{}, diags
		}

		authType, err := parseGitConnectionAuthType(gitConnectionModel.AuthType.ValueString())
		if err != nil {
			return []sifflet.GitConnection{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to create Looker source", err.Error()),
			}
		}
		gitConnections[i] = sifflet.GitConnection{
			AuthType: authType,
			Branch:   gitConnectionModel.Branch.ValueStringPointer(),
			SecretId: gitConnectionModel.SecretId.ValueString(),
			Url:      gitConnectionModel.Url.ValueString(),
		}
	}
	return gitConnections, diag.Diagnostics{}
}

func (m LookerParametersModel) ToCreateDto(ctx context.Context, name string) (sifflet.PublicCreateSourceV2JSONBody, diag.Diagnostics) {
	gitConnections, diags := makeGitConnectionsDto(ctx, m)
	if diags.HasError() {
		return sifflet.PublicCreateSourceV2JSONBody{}, diags
	}

	lookerInformation := sifflet.LookerInformation{
		GitConnections: &gitConnections,
		Host:           m.Host.ValueString(),
	}

	lookerCreateDto := &sifflet.PublicCreateLookerSourceV2Dto{
		Name:              name,
		Type:              sifflet.PublicCreateLookerSourceV2DtoTypeLOOKER,
		LookerInformation: &lookerInformation,
		Credentials:       m.Credentials.ValueStringPointer(),
		Schedule:          m.Schedule.ValueStringPointer(),
	}

	var createSourceJsonBody sifflet.PublicCreateSourceV2JSONBody
	err := createSourceJsonBody.FromAny(lookerCreateDto)
	if err != nil {
		return sifflet.PublicCreateSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot create Looker source", err)
	}

	return createSourceJsonBody, diag.Diagnostics{}
}

func (m LookerParametersModel) ToUpdateDto(ctx context.Context, name string) (sifflet.PublicEditSourceV2JSONBody, diag.Diagnostics) {
	gitConnections, diags := makeGitConnectionsDto(ctx, m)
	if diags.HasError() {
		return sifflet.PublicEditSourceV2JSONBody{}, diags
	}

	lookerInformation := sifflet.LookerInformation{
		GitConnections: &gitConnections,
		Host:           m.Host.ValueString(),
	}

	lookerUpdateDto := &sifflet.PublicUpdateLookerSourceV2Dto{
		Name:              &name,
		Type:              sifflet.PublicUpdateLookerSourceV2DtoTypeLOOKER,
		LookerInformation: lookerInformation,
		Credentials:       m.Credentials.ValueString(),
		Schedule:          m.Schedule.ValueStringPointer(),
	}

	var editSourceJsonBody sifflet.PublicEditSourceV2JSONBody
	err := editSourceJsonBody.FromAny(lookerUpdateDto)
	if err != nil {
		return sifflet.PublicEditSourceV2JSONBody{}, tfutils.ErrToDiags("Cannot update Looker source", err)
	}

	return editSourceJsonBody, diag.Diagnostics{}
}

func (m *LookerParametersModel) ModelFromDto(ctx context.Context, d sifflet.SiffletPublicGetSourceV2Dto) diag.Diagnostics {
	lookerDto := d.PublicGetLookerSourceV2Dto
	if lookerDto == nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic("Cannot read Looker source", "Source does not contain Looker params but was interpreted as a Looker source")}
	}

	// Convert git connections from DTO
	gitConnectionModels := make([]gitConnectionModel, len(*lookerDto.LookerInformation.GitConnections))
	for i, gitConnectionDto := range *lookerDto.LookerInformation.GitConnections {
		authType, err := gitConnectionAuthTypeToString(gitConnectionDto.AuthType)
		if err != nil {
			return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to read Looker source", err.Error())}
		}
		gitConnectionModels[i] = gitConnectionModel{
			AuthType: types.StringValue(authType),
			Branch:   types.StringPointerValue(gitConnectionDto.Branch),
			SecretId: types.StringValue(gitConnectionDto.SecretId),
			Url:      types.StringValue(gitConnectionDto.Url),
		}
	}

	gitConnections, diags := types.ListValueFrom(ctx, gitConnectionTypeAttributes, gitConnectionModels)
	if diags.HasError() {
		return diags
	}

	m.GitConnections = gitConnections
	m.Host = types.StringValue(lookerDto.LookerInformation.Host)
	m.Credentials = types.StringPointerValue(lookerDto.Credentials)
	m.Schedule = types.StringPointerValue(lookerDto.Schedule)
	return diag.Diagnostics{}
}

package source

import (
	"context"
	"fmt"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type LookerParametersModel struct {
	GitConnections types.List   `tfsdk:"git_connections"`
	Host           types.String `tfsdk:"host"`
}

type gitConnectionModel struct {
	AuthType types.String `tfsdk:"auth_type"`
	Branch   types.String `tfsdk:"branch"`
	SecretId types.String `tfsdk:"secret_id"`
	Url      types.String `tfsdk:"url"`
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
		},
	}
}

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

func (m LookerParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"git_connections": types.ListType{ElemType: gitConnectionTypeAttributes},
		"host":            types.StringType,
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

func (m LookerParametersModel) IsRepresentedBy(model ParametersModel) bool {
	return !model.Looker.IsNull()
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

func (m *LookerParametersModel) DtoFromModel(ctx context.Context, p ParametersModel) (sifflet.PublicCreateSourceDto_Parameters, diag.Diagnostics) {
	var parametersDto sifflet.PublicCreateSourceDto_Parameters
	diags := p.Looker.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}

	gitConnections := make([]types.Object, 0, len(m.GitConnections.Elements()))
	diags = m.GitConnections.ElementsAs(ctx, &gitConnections, false)
	if diags.HasError() {
		return sifflet.PublicCreateSourceDto_Parameters{}, diags
	}

	gitConnectionsDto := make([]sifflet.GitConnection, len(gitConnections))
	for i, gitConnection := range gitConnections {
		gitConnectionModel := gitConnectionModel{}
		diags = gitConnection.As(ctx, &gitConnectionModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return sifflet.PublicCreateSourceDto_Parameters{}, diags
		}
		authType, err := parseGitConnectionAuthType(gitConnectionModel.AuthType.ValueString())
		if err != nil {
			return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to create source", err.Error()),
			}
		}
		gitConnectionsDto[i] = sifflet.GitConnection{
			AuthType: authType,
			Branch:   gitConnectionModel.Branch.ValueStringPointer(),
			SecretId: gitConnectionModel.SecretId.ValueString(),
			Url:      gitConnectionModel.Url.ValueString(),
		}
	}

	dto := sifflet.PublicLookerParametersDto{
		GitConnections: &gitConnectionsDto,
		Host:           m.Host.ValueStringPointer(),
		Type:           sifflet.PublicLookerParametersDtoTypeLOOKER,
	}
	err := parametersDto.FromPublicLookerParametersDto(dto)
	if err != nil {
		return sifflet.PublicCreateSourceDto_Parameters{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to create source", err.Error()),
		}
	}
	return parametersDto, diag.Diagnostics{}
}

func (m *LookerParametersModel) ModelFromDto(ctx context.Context, d sifflet.PublicGetSourceDto_Parameters) diag.Diagnostics {
	paramsDto, err := d.AsPublicLookerParametersDto()
	if diags := handleDtoToModelError(err, m.SchemaSourceType()); diags.HasError() {
		return diags
	}
	m.Host = types.StringPointerValue(paramsDto.Host)

	gitConnectionModels := make([]gitConnectionModel, len(*paramsDto.GitConnections))
	for i, gitConnectionDto := range *paramsDto.GitConnections {
		authType, err := gitConnectionAuthTypeToString(gitConnectionDto.AuthType)
		if err != nil {
			return diag.Diagnostics{diag.NewErrorDiagnostic("Unable to read source", err.Error())}
		}
		gitConnectionModels[i] = gitConnectionModel{
			AuthType: types.StringValue(authType),
			Branch:   types.StringPointerValue(gitConnectionDto.Branch),
			SecretId: types.StringValue(gitConnectionDto.SecretId),
			Url:      types.StringValue(gitConnectionDto.Url),
		}
	}

	var diags diag.Diagnostics
	m.GitConnections, diags = types.ListValueFrom(ctx, gitConnectionTypeAttributes, gitConnectionModels)
	return diags
}

func (m LookerParametersModel) RequiresCredential() bool {
	return true
}

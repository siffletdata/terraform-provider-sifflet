package credentials

import (
	"context"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/model"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type credentialModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Value       types.String `tfsdk:"value"`
}

var (
	_ model.ModelWithId[string]                                            = credentialModel{}
	_ model.CreatableModel[sifflet.PublicCredentialsCreateDto]             = credentialModel{}
	_ model.UpdatableModel[sifflet.PublicUpdateCredentialsJSONRequestBody] = credentialModel{}
	// ReadableModel isn't implemented, since the API doesn't return the credentials value, so some custom
	// logic is used instead.
)

func (m credentialModel) ModelId() (string, diag.Diagnostics) {
	return m.Name.ValueString(), diag.Diagnostics{}
}

func (m credentialModel) ToCreateDto(_ context.Context) (sifflet.PublicCredentialsCreateDto, diag.Diagnostics) {
	return sifflet.PublicCredentialsCreateDto{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueStringPointer(),
		Value:       m.Value.ValueString(),
	}, diag.Diagnostics{}

}

func (m credentialModel) ToUpdateDto(_ context.Context) (sifflet.PublicUpdateCredentialsJSONRequestBody, diag.Diagnostics) {
	return sifflet.PublicUpdateCredentialsJSONRequestBody{
		Description: m.Description.ValueStringPointer(),
		Value:       m.Value.ValueStringPointer(),
	}, diag.Diagnostics{}
}

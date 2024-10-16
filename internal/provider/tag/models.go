package tag

import (
	"context"
	"terraform-provider-sifflet/internal/model"
	"terraform-provider-sifflet/internal/tfutils"

	sifflet "terraform-provider-sifflet/internal/alphaclient"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tagModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

var (
	_ model.FullModel[sifflet.TagDto, sifflet.TagCreateDto, sifflet.TagUpdateDto] = &tagModel{}
	_ model.ModelWithId[uuid.UUID]                                                = &tagModel{}
)

func (m *tagModel) ToCreateDto(_ context.Context) (sifflet.TagCreateDto, diag.Diagnostics) {
	return sifflet.TagCreateDto{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueStringPointer(),
		Type:        sifflet.TagCreateDtoTypeGENERIC,
	}, diag.Diagnostics{}
}

func (m tagModel) ToUpdateDto(_ context.Context) (sifflet.TagUpdateDto, diag.Diagnostics) {
	return sifflet.TagUpdateDto{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueStringPointer(),
	}, diag.Diagnostics{}
}

func (m *tagModel) FromDto(_ context.Context, dto sifflet.TagDto) diag.Diagnostics {
	m.Id = types.StringValue(dto.Id.String())
	m.Name = types.StringValue(dto.Name)
	m.Description = types.StringPointerValue(dto.Description)
	return diag.Diagnostics{}
}

func (m tagModel) ModelId() (uuid.UUID, diag.Diagnostics) {
	id, err := uuid.Parse(m.Id.ValueString())
	if err != nil {
		return uuid.Nil, tfutils.ErrToDiags("Could not parse ID as UUID", err)
	}
	return id, diag.Diagnostics{}
}

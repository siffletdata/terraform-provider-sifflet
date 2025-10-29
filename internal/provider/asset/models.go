package asset

import (
	"context"
	"terraform-provider-sifflet/internal/model"
	"terraform-provider-sifflet/internal/provider/tag"
	"terraform-provider-sifflet/internal/tfutils"

	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type assetModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Uri         types.String `tfsdk:"uri"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
}

var (
	_ model.ReadableModel[sifflet.PublicGetAssetDto] = &assetModel{}
	_ model.ModelWithId[uuid.UUID]                   = &assetModel{}
)

func (m assetModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
		"type":        types.StringType,
		"uri":         types.StringType,
		"tags": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: tag.PublicApiTagModel{}.AttributeTypes(),
			},
		},
	}
}

func (m *assetModel) FromDto(ctx context.Context, dto sifflet.PublicGetAssetDto) diag.Diagnostics {
	tags, diags := model.NewModelListFromDto(ctx, *dto.Tags,
		func() model.InnerModel[sifflet.PublicTagReferenceDto] { return &tag.PublicApiTagModel{} },
	)
	if diags.HasError() {
		return diags
	}

	m.Id = types.StringValue(dto.Id.String())
	m.Name = types.StringValue(dto.Name)
	m.Description = types.StringPointerValue(dto.Description)
	m.Type = types.StringValue(string(dto.Type))
	m.Uri = types.StringValue(dto.Uri)
	m.Tags = tags
	return diag.Diagnostics{}
}

func (m *assetModel) FromListDto(ctx context.Context, dto sifflet.PublicGetAssetListDto) diag.Diagnostics {
	tags, diags := model.NewModelListFromDto(ctx, *dto.Tags,
		func() model.InnerModel[sifflet.PublicTagReferenceDto] { return &tag.PublicApiTagModel{} },
	)
	if diags.HasError() {
		return diags
	}

	m.Id = types.StringValue(dto.Id.String())
	m.Name = types.StringValue(dto.Name)
	m.Description = types.StringPointerValue(dto.Description)
	m.Type = types.StringValue(string(dto.Type))
	m.Uri = types.StringValue(dto.Uri)
	m.Tags = tags
	return diag.Diagnostics{}
}

func (m assetModel) ModelId() (uuid.UUID, diag.Diagnostics) {
	id, err := uuid.Parse(m.Id.ValueString())
	if err != nil {
		return uuid.Nil, tfutils.ErrToDiags("Could not parse ID as UUID", err)
	}
	return id, diag.Diagnostics{}
}

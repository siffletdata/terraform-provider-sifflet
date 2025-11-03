package domain

import (
	"context"

	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/model"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type staticContentDefinitionModel struct {
	AssetUris types.Set `tfsdk:"asset_uris"`
}

type tagModel struct {
	Kind types.String `tfsdk:"kind"`
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

type dynamicContentDefinitionConditionModel struct {
	LogicalOperator types.String `tfsdk:"logical_operator"`
	SchemaUris      types.List   `tfsdk:"schema_uris"`
	Tags            types.List   `tfsdk:"tags"`
}

type dynamicContentDefinitionModel struct {
	LogicalOperator types.String `tfsdk:"logical_operator"`
	Conditions      types.List   `tfsdk:"conditions"`
}

type domainModel struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	DynamicContentDefinition types.Object `tfsdk:"dynamic_content_definition"`
	StaticContentDefinition  types.Object `tfsdk:"static_content_definition"`
}

var (
	_ model.FullModel[sifflet.PublicGetDomainDto, sifflet.PublicCreateDomainDto, sifflet.PublicUpdateDomainDto] = &domainModel{}
	_ model.ModelWithId[uuid.UUID]                                                                              = &domainModel{}
)

func (m domainModel) ModelId() (uuid.UUID, diag.Diagnostics) {
	id, err := uuid.Parse(m.Id.ValueString())
	if err != nil {
		return uuid.Nil, tfutils.ErrToDiags("Could not parse ID as UUID", err)
	}
	return id, diag.Diagnostics{}
}

func (m domainModel) getDynamicContentDefinitionDto(ctx context.Context) (sifflet.PublicDynamicDomainContentDefinitionDto, diag.Diagnostics) {
	var dynamicContentDefinitionModel dynamicContentDefinitionModel
	diags := m.DynamicContentDefinition.As(ctx, &dynamicContentDefinitionModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicDynamicDomainContentDefinitionDto{}, diags
	}

	conditionsModel := make([]types.Object, 0, len(dynamicContentDefinitionModel.Conditions.Elements()))
	diags = dynamicContentDefinitionModel.Conditions.ElementsAs(ctx, &conditionsModel, false)
	if diags.HasError() {
		return sifflet.PublicDynamicDomainContentDefinitionDto{}, diags
	}
	conditions := make([]sifflet.PublicDynamicDomainContentDefinitionDto_Conditions_Item, 0, len(dynamicContentDefinitionModel.Conditions.Elements()))
	for _, condition := range conditionsModel {
		var conditionModel dynamicContentDefinitionConditionModel
		diags := condition.As(ctx, &conditionModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return sifflet.PublicDynamicDomainContentDefinitionDto{}, diags
		}
		conditionDto, diags := conditionModel.ToDto(ctx)
		if diags.HasError() {
			return sifflet.PublicDynamicDomainContentDefinitionDto{}, diags
		}
		conditions = append(conditions, conditionDto)
	}
	filterLogicalOperator := sifflet.PublicDynamicDomainContentDefinitionDtoFilterLogicalOperator(dynamicContentDefinitionModel.LogicalOperator.ValueString())
	return sifflet.PublicDynamicDomainContentDefinitionDto{
		Type:                  sifflet.PublicDynamicDomainContentDefinitionDtoTypeDYNAMIC,
		FilterLogicalOperator: &filterLogicalOperator,
		Conditions:            &conditions,
	}, diag.Diagnostics{}
}

func (m domainModel) getStaticContentDefinitionDto(ctx context.Context) (sifflet.PublicStaticDomainContentDefinitionDto, diag.Diagnostics) {
	var staticContentDefinitionModel staticContentDefinitionModel
	diags := m.StaticContentDefinition.As(ctx, &staticContentDefinitionModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return sifflet.PublicStaticDomainContentDefinitionDto{}, diags
	}

	assets := make([]string, 0, len(staticContentDefinitionModel.AssetUris.Elements()))
	assetsModel := make([]types.String, 0, len(staticContentDefinitionModel.AssetUris.Elements()))
	diags = staticContentDefinitionModel.AssetUris.ElementsAs(ctx, &assetsModel, false)
	if diags.HasError() {
		return sifflet.PublicStaticDomainContentDefinitionDto{}, diags
	}
	for _, assetUri := range assetsModel {
		assets = append(assets, assetUri.ValueString())
	}

	return sifflet.PublicStaticDomainContentDefinitionDto{
		Type:   sifflet.STATIC,
		Assets: &assets,
	}, diag.Diagnostics{}
}

func (m domainModel) ToCreateDto(ctx context.Context) (sifflet.PublicCreateDomainDto, diag.Diagnostics) {
	var assetContentDefinition sifflet.PublicCreateDomainDto_AssetContentDefinition

	// Dynamic content definition
	if !m.DynamicContentDefinition.IsNull() && !m.DynamicContentDefinition.IsUnknown() {
		dymanicContentDefinition, diags := m.getDynamicContentDefinitionDto(ctx)
		if diags.HasError() {
			return sifflet.PublicCreateDomainDto{}, diags
		}
		err := assetContentDefinition.FromPublicDynamicDomainContentDefinitionDto(dymanicContentDefinition)
		if err != nil {
			return sifflet.PublicCreateDomainDto{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to create domain", err.Error()),
			}
		}
	}

	// Static content definition
	if !m.StaticContentDefinition.IsNull() && !m.StaticContentDefinition.IsUnknown() {
		staticContentDefintion, diags := m.getStaticContentDefinitionDto(ctx)
		if diags.HasError() {
			return sifflet.PublicCreateDomainDto{}, diags
		}
		err := assetContentDefinition.FromPublicStaticDomainContentDefinitionDto(staticContentDefintion)
		if err != nil {
			return sifflet.PublicCreateDomainDto{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to create domain", err.Error()),
			}
		}
	}

	return sifflet.PublicCreateDomainDto{
		Name:                   m.Name.ValueString(),
		Description:            m.Description.ValueStringPointer(),
		AssetContentDefinition: assetContentDefinition,
	}, diag.Diagnostics{}
}

func (m domainModel) ToUpdateDto(ctx context.Context) (sifflet.PublicUpdateDomainDto, diag.Diagnostics) {
	var assetContentDefinition sifflet.PublicUpdateDomainDto_AssetContentDefinition

	// Dynamic content definition
	if !m.DynamicContentDefinition.IsNull() && !m.DynamicContentDefinition.IsUnknown() {
		dymanicContentDefinition, diags := m.getDynamicContentDefinitionDto(ctx)
		if diags.HasError() {
			return sifflet.PublicUpdateDomainDto{}, diags
		}
		err := assetContentDefinition.FromPublicDynamicDomainContentDefinitionDto(dymanicContentDefinition)
		if err != nil {
			return sifflet.PublicUpdateDomainDto{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to create domain", err.Error()),
			}
		}
	}

	// Static content definition
	if !m.StaticContentDefinition.IsNull() && !m.StaticContentDefinition.IsUnknown() {
		staticContentDefintion, diags := m.getStaticContentDefinitionDto(ctx)
		if diags.HasError() {
			return sifflet.PublicUpdateDomainDto{}, diags
		}
		err := assetContentDefinition.FromPublicStaticDomainContentDefinitionDto(staticContentDefintion)
		if err != nil {
			return sifflet.PublicUpdateDomainDto{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to create domain", err.Error()),
			}
		}
	}

	return sifflet.PublicUpdateDomainDto{
		Name:                   m.Name.ValueString(),
		Description:            m.Description.ValueStringPointer(),
		AssetContentDefinition: assetContentDefinition,
	}, diag.Diagnostics{}
}

func (m *domainModel) FromDto(ctx context.Context, dto sifflet.PublicGetDomainDto) diag.Diagnostics {
	contentDefinitionType, err := sifflet.GetDomainContentDefinitionType(dto.AssetContentDefinition)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to read domain content definition type", err.Error()),
		}
	}

	var dynamicContentDefinitionModel dynamicContentDefinitionModel
	var staticContentDefinitionModel staticContentDefinitionModel

	if contentDefinitionType == string(sifflet.PublicDomainContentDefinitionDtoTypeDYNAMIC) {
		dynamicContentDefinition, err := dto.AssetContentDefinition.AsPublicDynamicDomainContentDefinitionDto()
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to read domain content definition", err.Error()),
			}
		}
		diags := dynamicContentDefinitionModel.FromDto(ctx, dynamicContentDefinition)
		if diags.HasError() {
			return diags
		}
		dynamicContentDefinitionObject, diags := types.ObjectValueFrom(ctx, dynamicContentDefinitionModel.AttributeTypes(), dynamicContentDefinitionModel)
		if diags.HasError() {
			return diags
		}
		m.DynamicContentDefinition = dynamicContentDefinitionObject
		m.StaticContentDefinition = types.ObjectNull(staticContentDefinitionModel.AttributeTypes())
	}

	if contentDefinitionType == string(sifflet.PublicDomainContentDefinitionDtoTypeSTATIC) {
		staticContentDefinition, err := dto.AssetContentDefinition.AsPublicStaticDomainContentDefinitionDto()
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to read domain content definition", err.Error()),
			}
		}
		diags := staticContentDefinitionModel.FromDto(ctx, staticContentDefinition)
		if diags.HasError() {
			return diags
		}
		staticContentDefinitionObject, diags := types.ObjectValueFrom(ctx, staticContentDefinitionModel.AttributeTypes(), staticContentDefinitionModel)
		if diags.HasError() {
			return diags
		}
		m.StaticContentDefinition = staticContentDefinitionObject
		m.DynamicContentDefinition = types.ObjectNull(dynamicContentDefinitionModel.AttributeTypes())
	}

	m.Id = types.StringValue(dto.Id.String())
	m.Name = types.StringValue(dto.Name)
	m.Description = types.StringPointerValue(dto.Description)

	return diag.Diagnostics{}
}

// Static content definition model.
func (m staticContentDefinitionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"asset_uris": types.SetType{
			ElemType: types.StringType,
		},
	}
}

func (m *staticContentDefinitionModel) FromDto(ctx context.Context, dto sifflet.PublicStaticDomainContentDefinitionDto) diag.Diagnostics {
	assets := make([]types.String, 0, len(*dto.Assets))
	for _, asset := range *dto.Assets {
		assets = append(assets, types.StringValue(asset))
	}
	assetUris, diags := types.SetValueFrom(ctx, types.StringType, assets)
	if diags.HasError() {
		return diags
	}
	m.AssetUris = assetUris
	return diags
}

// Dynamic content definition model.
func (m dynamicContentDefinitionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_operator": types.StringType,
		"conditions": types.ListType{
			ElemType: types.ObjectType{AttrTypes: dynamicContentDefinitionConditionModel{}.AttributeTypes()},
		},
	}
}

func (m *dynamicContentDefinitionModel) FromDto(ctx context.Context, dto sifflet.PublicDynamicDomainContentDefinitionDto) diag.Diagnostics {
	logicalOperator := types.StringValue(string(*dto.FilterLogicalOperator))

	conditionList, diags := model.NewModelListFromDto(
		ctx, *dto.Conditions,
		func() model.InnerModel[sifflet.PublicDynamicDomainContentDefinitionDto_Conditions_Item] {
			return &dynamicContentDefinitionConditionModel{}
		},
	)
	if diags.HasError() {
		return diags
	}

	m.LogicalOperator = logicalOperator
	m.Conditions = conditionList
	return diag.Diagnostics{}
}

// Dynamic content definition condition model.
func (m dynamicContentDefinitionConditionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logical_operator": types.StringType,
		"schema_uris": types.ListType{
			ElemType: types.StringType,
		},
		"tags": types.ListType{
			ElemType: types.ObjectType{AttrTypes: tagModel{}.AttributeTypes()},
		},
	}
}

func (m *dynamicContentDefinitionConditionModel) FromDto(ctx context.Context, dto sifflet.PublicDynamicDomainContentDefinitionDto_Conditions_Item) diag.Diagnostics {
	conditionType, err := sifflet.GetDomainDynamicDomainConditionType(dto)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to read domain condition type", err.Error()),
		}
	}

	if conditionType == string(sifflet.PublicFilterDomainConditionDtoTypeSOURCE) {
		sourceFilterCondition, err := dto.AsPublicSourceFilterDomainConditionDto()
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to read domain condition", err.Error()),
			}
		}
		schemaUris, diags := types.ListValueFrom(ctx, types.StringType, sourceFilterCondition.Sources)
		if diags.HasError() {
			return diags
		}
		m.LogicalOperator = types.StringValue(string(*sourceFilterCondition.Operator))
		m.SchemaUris = schemaUris
		m.Tags = types.ListNull(types.ObjectType{AttrTypes: tagModel{}.AttributeTypes()})
	}

	if conditionType == string(sifflet.PublicFilterDomainConditionDtoTypeTAG) {
		tagFilterCondition, err := dto.AsPublicTagFilterDomainConditionDto()
		if err != nil {
			return diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to read domain condition", err.Error()),
			}
		}
		tags, diags := model.NewModelListFromDto(
			ctx, *tagFilterCondition.Tags,
			func() model.InnerModel[sifflet.PublicExternalTagReferenceDto] { return &tagModel{} },
		)
		if diags.HasError() {
			return diags
		}
		m.LogicalOperator = types.StringValue(string(*tagFilterCondition.Operator))
		m.Tags = tags
		m.SchemaUris = types.ListNull(types.StringType)
	}
	return diag.Diagnostics{}
}

func (m dynamicContentDefinitionConditionModel) ToDto(ctx context.Context) (sifflet.PublicDynamicDomainContentDefinitionDto_Conditions_Item, diag.Diagnostics) {
	var conditionDto sifflet.PublicDynamicDomainContentDefinitionDto_Conditions_Item
	// Source filter condition
	if !m.SchemaUris.IsNull() && !m.SchemaUris.IsUnknown() {
		schemaUris := make([]string, 0, len(m.SchemaUris.Elements()))

		schemaUrisModel := make([]types.String, 0, len(m.SchemaUris.Elements()))
		diags := m.SchemaUris.ElementsAs(ctx, &schemaUrisModel, false)
		if diags.HasError() {
			return sifflet.PublicDynamicDomainContentDefinitionDto_Conditions_Item{}, diags
		}
		for _, schemaUri := range schemaUrisModel {
			schemaUris = append(schemaUris, schemaUri.ValueString())
		}
		operator := sifflet.PublicSourceFilterDomainConditionDtoOperator(m.LogicalOperator.ValueString())
		sourceFilterCondition := sifflet.PublicSourceFilterDomainConditionDto{
			Type:     sifflet.PublicSourceFilterDomainConditionDtoTypeSOURCE,
			Sources:  &schemaUris,
			Operator: &operator,
		}
		err := conditionDto.FromPublicSourceFilterDomainConditionDto(sourceFilterCondition)
		if err != nil {
			return sifflet.PublicDynamicDomainContentDefinitionDto_Conditions_Item{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to create domain", err.Error()),
			}
		}
	}

	// Tag filter condition
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		tags := make([]sifflet.PublicExternalTagReferenceDto, 0, len(m.Tags.Elements()))
		tagsModel := make([]types.Object, 0, len(m.Tags.Elements()))
		diags := m.Tags.ElementsAs(ctx, &tagsModel, false)
		if diags.HasError() {
			return sifflet.PublicDynamicDomainContentDefinitionDto_Conditions_Item{}, diags
		}
		for _, tag := range tagsModel {
			var tagModel tagModel
			diags := tag.As(ctx, &tagModel, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return sifflet.PublicDynamicDomainContentDefinitionDto_Conditions_Item{}, diags
			}
			tagDto, diags := tagModel.ToDto(ctx)
			if diags.HasError() {
				return sifflet.PublicDynamicDomainContentDefinitionDto_Conditions_Item{}, diags
			}
			tags = append(tags, tagDto)
		}
		operator := sifflet.PublicTagFilterDomainConditionDtoOperator(m.LogicalOperator.ValueString())
		tagFilterCondition := sifflet.PublicTagFilterDomainConditionDto{
			Type:     sifflet.PublicTagFilterDomainConditionDtoTypeTAG,
			Tags:     &tags,
			Operator: &operator,
		}
		err := conditionDto.FromPublicTagFilterDomainConditionDto(tagFilterCondition)
		if err != nil {
			return sifflet.PublicDynamicDomainContentDefinitionDto_Conditions_Item{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Unable to create domain", err.Error()),
			}
		}
	}
	return conditionDto, diag.Diagnostics{}
}

// Tag model.
func (m tagModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"kind": types.StringType,
		"name": types.StringType,
		"id":   types.StringType,
	}
}

func (m *tagModel) FromDto(ctx context.Context, dto sifflet.PublicExternalTagReferenceDto) diag.Diagnostics {
	m.Kind = types.StringValue(string(*dto.Kind))
	m.Name = types.StringValue(*dto.Name)
	m.Id = types.StringValue(dto.Id.String())
	return diag.Diagnostics{}
}

func (m tagModel) ToDto(_ context.Context) (sifflet.PublicExternalTagReferenceDto, diag.Diagnostics) {
	var id *uuid.UUID
	var kind *sifflet.PublicExternalTagReferenceDtoKind
	var name *string
	if !m.Id.IsNull() && m.Id.ValueString() != "" {
		// If an ID was provided, the DTO should not include a name or kind
		idv, err := uuid.Parse(m.Id.ValueString())
		if err != nil {
			return sifflet.PublicExternalTagReferenceDto{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Tag ID is not a valid UUID", err.Error()),
			}
		}
		id = &idv
	} else {
		// If an ID is not provided, then a name was provided (enforced by the schema)
		// Let's double check that here for clarity.
		if m.Name.IsNull() || m.Name.ValueString() == "" {
			return sifflet.PublicExternalTagReferenceDto{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Tag name is required when an ID is not provided", ""),
			}
		}
		name = m.Name.ValueStringPointer()
		if !m.Kind.IsNull() && m.Kind.ValueString() != "" {
			t := sifflet.PublicExternalTagReferenceDtoKind(m.Kind.ValueString())
			kind = &t
		}
	}

	return sifflet.PublicExternalTagReferenceDto{
		Id:   id,
		Name: name,
		Kind: kind,
	}, diag.Diagnostics{}
}

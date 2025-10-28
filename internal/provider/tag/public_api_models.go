package tag

import (
	"context"
	"terraform-provider-sifflet/internal/model"

	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PublicApiTagModel struct {
	Name types.String `tfsdk:"name"`
	ID   types.String `tfsdk:"id"`
	Kind types.String `tfsdk:"kind"`
}

var (
	_ model.InnerModel[sifflet.PublicTagReferenceDto] = &PublicApiTagModel{}
)

func (m PublicApiTagModel) ToDto() (sifflet.PublicTagReferenceDto, diag.Diagnostics) {
	var id *uuid.UUID
	var kind *sifflet.PublicTagReferenceDtoKind
	var name *string
	if !m.ID.IsNull() && m.ID.ValueString() != "" {
		// If an ID was provided, the DTO should not include a name or kind
		idv, err := uuid.Parse(m.ID.ValueString())
		if err != nil {
			return sifflet.PublicTagReferenceDto{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Tag ID is not a valid UUID", err.Error()),
			}
		}
		id = &idv
	} else {
		// If an ID is not provided, then a name was provided (enforced by the schema)
		// Let's double check that here for clarity.
		if m.Name.IsNull() || m.Name.ValueString() == "" {
			return sifflet.PublicTagReferenceDto{}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Tag name is required when an ID is not provided", ""),
			}
		}
		name = m.Name.ValueStringPointer()
		if !m.Kind.IsNull() && m.Kind.ValueString() != "" {
			t := sifflet.PublicTagReferenceDtoKind(m.Kind.ValueString())
			kind = &t
		}
	}

	return sifflet.PublicTagReferenceDto{
		Id:   id,
		Name: name,
		Kind: kind,
	}, diag.Diagnostics{}
}

func (m *PublicApiTagModel) FromDto(_ context.Context, dto sifflet.PublicTagReferenceDto) diag.Diagnostics {
	m.ID = types.StringValue(dto.Id.String())
	m.Name = types.StringPointerValue(dto.Name)
	kind := "Tag"
	if dto.Kind != nil {
		kind = string(*dto.Kind)
	}
	// Backwards compatibility. At some point, the API
	// started returning "TAG", but the resource schema
	// expects "Tag".
	if kind == "TAG" {
		kind = "Tag"
	}
	if kind == "CLASSIFICATION" {
		kind = "Classification"
	}
	m.Kind = types.StringValue(kind)
	return diag.Diagnostics{}
}

func (m PublicApiTagModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"id":   types.StringType,
		"kind": types.StringType,
	}
}

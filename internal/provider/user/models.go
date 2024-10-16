package user

import (
	"context"

	"terraform-provider-sifflet/internal/client"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/model"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type userModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Email       types.String `tfsdk:"email"`
	Role        types.String `tfsdk:"role"`
	Permissions types.List   `tfsdk:"permissions"`
}

var (
	_ model.FullModel[sifflet.PublicUserGetDto, sifflet.PublicUserCreateDto, sifflet.PublicUserUpdateDto] = &userModel{}
	_ model.ModelWithId[uuid.UUID]                                                                        = userModel{}
)

func (m userModel) ModelId() (uuid.UUID, diag.Diagnostics) {
	id, err := uuid.Parse(m.Id.ValueString())
	if err != nil {
		return uuid.Nil, tfutils.ErrToDiags("Could not parse ID as UUID", err)
	}
	return id, diag.Diagnostics{}
}

func (m userModel) getPermissionsModel() ([]permissionModel, diag.Diagnostics) {
	permissions := make([]permissionModel, 0, len(m.Permissions.Elements()))
	diags := m.Permissions.ElementsAs(context.Background(), &permissions, false)
	return permissions, diags
}

func (m userModel) ToCreateDto(_ context.Context) (sifflet.PublicUserCreateDto, diag.Diagnostics) {
	permissionsModel, diags := m.getPermissionsModel()
	if diags.HasError() {
		return sifflet.PublicUserCreateDto{}, diags
	}

	permissionsDto := make([]sifflet.PublicUserPermissionAssignmentDto, len(permissionsModel))
	for i, permissionModel := range permissionsModel {
		dto, diags := permissionModel.ToDto()
		if diags.HasError() {
			return sifflet.PublicUserCreateDto{}, diags
		}
		permissionsDto[i] = dto
	}

	return sifflet.PublicUserCreateDto{
		Email:       m.Email.ValueString(),
		Name:        m.Name.ValueString(),
		Role:        sifflet.PublicUserCreateDtoRole(m.Role.ValueString()),
		Permissions: permissionsDto,
	}, diag.Diagnostics{}

}

func (m userModel) ToUpdateDto(_ context.Context) (sifflet.PublicUserUpdateDto, diag.Diagnostics) {
	permissionsModel, diags := m.getPermissionsModel()
	if diags.HasError() {
		return sifflet.PublicUserUpdateDto{}, diags
	}

	permissionsDto := make([]sifflet.PublicUserPermissionAssignmentDto, len(permissionsModel))
	for i, permissionModel := range permissionsModel {
		dto, diags := permissionModel.ToDto()
		if diags.HasError() {
			return sifflet.PublicUserUpdateDto{}, diags
		}
		permissionsDto[i] = dto
	}

	return sifflet.PublicUserUpdateDto{
		Name:        m.Name.ValueString(),
		Role:        sifflet.PublicUserUpdateDtoRole(m.Role.ValueString()),
		Permissions: permissionsDto,
	}, diag.Diagnostics{}
}

func (m *userModel) FromDto(ctx context.Context, userDto sifflet.PublicUserGetDto) diag.Diagnostics {
	permissionsList, diags := model.NewModelListFromDto(
		ctx, userDto.Permissions,
		func() model.InnerModel[client.PublicUserPermissionAssignmentDto] { return &permissionModel{} },
	)
	if diags.HasError() {
		return diags
	}

	m.Id = types.StringValue(userDto.Id.String())
	m.Name = types.StringValue(userDto.Name)
	m.Email = types.StringValue(userDto.Email)
	m.Role = types.StringValue(string(userDto.Role))
	m.Permissions = permissionsList
	return diag.Diagnostics{}
}

var (
	_ model.InnerModel[sifflet.PublicUserPermissionAssignmentDto] = &permissionModel{}
)

type permissionModel struct {
	DomainId   types.String `tfsdk:"domain_id"`
	DomainRole types.String `tfsdk:"domain_role"`
}

func (m permissionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"domain_id":   types.StringType,
		"domain_role": types.StringType,
	}
}

func (m *permissionModel) FromDto(_ context.Context, dto sifflet.PublicUserPermissionAssignmentDto) diag.Diagnostics {
	domainRole := "VIEWER"
	if dto.DomainRole != nil {
		domainRole = string(*dto.DomainRole)
	}
	m.DomainId = types.StringValue(dto.DomainId.String())
	m.DomainRole = types.StringValue(domainRole)
	return diag.Diagnostics{}
}

func (m permissionModel) ToDto() (sifflet.PublicUserPermissionAssignmentDto, diag.Diagnostics) {
	uid, err := uuid.Parse(m.DomainId.ValueString())
	if err != nil {
		return sifflet.PublicUserPermissionAssignmentDto{}, tfutils.ErrToDiags("Could not parse domain ID as UUID", err)
	}
	if err != nil {
		return sifflet.PublicUserPermissionAssignmentDto{}, tfutils.ErrToDiags("Invalid domain role", err)
	}
	role := sifflet.PublicUserPermissionAssignmentDtoDomainRole(m.DomainRole.ValueString())
	return sifflet.PublicUserPermissionAssignmentDto{
		DomainId:   uid,
		DomainRole: &role,
	}, diag.Diagnostics{}
}

package team

import (
	"context"

	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/model"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type teamModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	DomainPermissions types.Set    `tfsdk:"domain_permissions"`
	Users             types.Set    `tfsdk:"users"`
}

var (
	_ model.FullModel[sifflet.PublicGetTeamDto, sifflet.PublicCreateTeamDto, sifflet.PublicUpdateTeamDto] = &teamModel{}
	_ model.ModelWithId[uuid.UUID]                                                                        = teamModel{}
)

func (m teamModel) ModelId() (uuid.UUID, diag.Diagnostics) {
	id, err := uuid.Parse(m.Id.ValueString())
	if err != nil {
		return uuid.Nil, tfutils.ErrToDiags("Could not parse ID as UUID", err)
	}
	return id, diag.Diagnostics{}
}

func (m teamModel) getDomainPermissionsModel() ([]domainPermissionModel, diag.Diagnostics) {
	if m.DomainPermissions.IsNull() || m.DomainPermissions.IsUnknown() {
		return []domainPermissionModel{}, diag.Diagnostics{}
	}
	permissions := make([]domainPermissionModel, 0, len(m.DomainPermissions.Elements()))
	diags := m.DomainPermissions.ElementsAs(context.Background(), &permissions, false)
	return permissions, diags
}

func (m teamModel) getUsersModel() ([]userReferenceModel, diag.Diagnostics) {
	if m.Users.IsNull() || m.Users.IsUnknown() {
		return []userReferenceModel{}, diag.Diagnostics{}
	}
	users := make([]userReferenceModel, 0, len(m.Users.Elements()))
	diags := m.Users.ElementsAs(context.Background(), &users, false)
	return users, diags
}

func (m teamModel) ToCreateDto(ctx context.Context) (sifflet.PublicCreateTeamDto, diag.Diagnostics) {
	domainPermissionsModel, diags := m.getDomainPermissionsModel()
	if diags.HasError() {
		return sifflet.PublicCreateTeamDto{}, diags
	}

	var domainPermissionsDto *[]sifflet.PublicTeamPermissionAssignmentDto
	if len(domainPermissionsModel) > 0 {
		permissions := make([]sifflet.PublicTeamPermissionAssignmentDto, len(domainPermissionsModel))
		for i, permissionModel := range domainPermissionsModel {
			dto, diags := permissionModel.ToDto(ctx)
			if diags.HasError() {
				return sifflet.PublicCreateTeamDto{}, diags
			}
			permissions[i] = dto
		}
		domainPermissionsDto = &permissions
	}

	usersModel, diags := m.getUsersModel()
	if diags.HasError() {
		return sifflet.PublicCreateTeamDto{}, diags
	}

	var usersDto *[]sifflet.PublicReferenceByIdOrEmailDto
	if len(usersModel) > 0 {
		users := make([]sifflet.PublicReferenceByIdOrEmailDto, len(usersModel))
		for i, userModel := range usersModel {
			dto, diags := userModel.ToDto(ctx)
			if diags.HasError() {
				return sifflet.PublicCreateTeamDto{}, diags
			}
			users[i] = dto
		}
		usersDto = &users
	}

	var description *string
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		desc := m.Description.ValueString()
		description = &desc
	}

	return sifflet.PublicCreateTeamDto{
		Name:              m.Name.ValueString(),
		Description:       description,
		DomainPermissions: domainPermissionsDto,
		Users:             usersDto,
	}, diag.Diagnostics{}
}

func (m teamModel) ToUpdateDto(ctx context.Context) (sifflet.PublicUpdateTeamDto, diag.Diagnostics) {
	domainPermissionsModel, diags := m.getDomainPermissionsModel()
	if diags.HasError() {
		return sifflet.PublicUpdateTeamDto{}, diags
	}

	var domainPermissionsDto *[]sifflet.PublicTeamPermissionAssignmentDto
	if !m.DomainPermissions.IsNull() && !m.DomainPermissions.IsUnknown() {
		permissions := make([]sifflet.PublicTeamPermissionAssignmentDto, len(domainPermissionsModel))
		for i, permissionModel := range domainPermissionsModel {
			dto, diags := permissionModel.ToDto(ctx)
			if diags.HasError() {
				return sifflet.PublicUpdateTeamDto{}, diags
			}
			permissions[i] = dto
		}
		domainPermissionsDto = &permissions
	}

	usersModel, diags := m.getUsersModel()
	if diags.HasError() {
		return sifflet.PublicUpdateTeamDto{}, diags
	}

	var usersDto *[]sifflet.PublicReferenceByIdOrEmailDto
	if !m.Users.IsNull() && !m.Users.IsUnknown() {
		users := make([]sifflet.PublicReferenceByIdOrEmailDto, len(usersModel))
		for i, userModel := range usersModel {
			dto, diags := userModel.ToDto(ctx)
			if diags.HasError() {
				return sifflet.PublicUpdateTeamDto{}, diags
			}
			users[i] = dto
		}
		usersDto = &users
	}

	var description *string
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		desc := m.Description.ValueString()
		description = &desc
	}

	return sifflet.PublicUpdateTeamDto{
		Name:              m.Name.ValueString(),
		Description:       description,
		DomainPermissions: domainPermissionsDto,
		Users:             usersDto,
	}, diag.Diagnostics{}
}

func (m *teamModel) FromDto(ctx context.Context, teamDto sifflet.PublicGetTeamDto) diag.Diagnostics {
	var domainPermissionsList types.Set
	if teamDto.DomainPermissions != nil && len(*teamDto.DomainPermissions) > 0 {
		permissionsList, diags := model.NewModelSetFromDto(
			ctx, *teamDto.DomainPermissions,
			func() model.InnerModel[sifflet.PublicTeamPermissionAssignmentDto] { return &domainPermissionModel{} },
		)
		if diags.HasError() {
			return diags
		}
		domainPermissionsList = permissionsList
	} else {
		// Return empty set instead of null to match Terraform's expectations
		var diags diag.Diagnostics
		domainPermissionsList, diags = types.SetValueFrom(ctx, types.ObjectType{AttrTypes: domainPermissionModel{}.AttributeTypes()}, []domainPermissionModel{})
		if diags.HasError() {
			return diags
		}
	}

	var usersList types.Set
	if teamDto.Users != nil && len(*teamDto.Users) > 0 {
		users, diags := model.NewModelSetFromDto(
			ctx, *teamDto.Users,
			func() model.InnerModel[sifflet.PublicReferenceByIdOrEmailDto] { return &userReferenceModel{} },
		)
		if diags.HasError() {
			return diags
		}
		usersList = users
	} else {
		// Return empty set instead of null to match Terraform's expectations
		var diags diag.Diagnostics
		usersList, diags = types.SetValueFrom(ctx, types.ObjectType{AttrTypes: userReferenceModel{}.AttributeTypes()}, []userReferenceModel{})
		if diags.HasError() {
			return diags
		}
	}

	m.Id = types.StringValue(teamDto.Id.String())
	m.Name = types.StringValue(teamDto.Name)
	if teamDto.Description != nil {
		m.Description = types.StringValue(*teamDto.Description)
	} else {
		m.Description = types.StringNull()
	}
	m.DomainPermissions = domainPermissionsList
	m.Users = usersList
	return diag.Diagnostics{}
}

var (
	_ model.InnerModel[sifflet.PublicTeamPermissionAssignmentDto] = &domainPermissionModel{}
)

type domainPermissionModel struct {
	DomainId   types.String `tfsdk:"domain_id"`
	DomainRole types.String `tfsdk:"domain_role"`
}

func (m domainPermissionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"domain_id":   types.StringType,
		"domain_role": types.StringType,
	}
}

func (m *domainPermissionModel) FromDto(_ context.Context, dto sifflet.PublicTeamPermissionAssignmentDto) diag.Diagnostics {
	m.DomainId = types.StringValue(dto.DomainId.String())
	m.DomainRole = types.StringValue(string(dto.DomainRole))
	return diag.Diagnostics{}
}

func (m domainPermissionModel) ToDto(_ context.Context) (sifflet.PublicTeamPermissionAssignmentDto, diag.Diagnostics) {
	uid, err := uuid.Parse(m.DomainId.ValueString())
	if err != nil {
		return sifflet.PublicTeamPermissionAssignmentDto{}, tfutils.ErrToDiags("Could not parse domain ID as UUID", err)
	}
	role := sifflet.PublicTeamPermissionAssignmentDtoDomainRole(m.DomainRole.ValueString())
	return sifflet.PublicTeamPermissionAssignmentDto{
		DomainId:   uid,
		DomainRole: role,
	}, diag.Diagnostics{}
}

var (
	_ model.InnerModel[sifflet.PublicReferenceByIdOrEmailDto] = &userReferenceModel{}
)

type userReferenceModel struct {
	UserId types.String `tfsdk:"user_id"`
	Email  types.String `tfsdk:"email"`
}

func (m userReferenceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"user_id": types.StringType,
		"email":   types.StringType,
	}
}

func (m *userReferenceModel) FromDto(_ context.Context, dto sifflet.PublicReferenceByIdOrEmailDto) diag.Diagnostics {
	if dto.Id != nil {
		m.UserId = types.StringValue(dto.Id.String())
	} else {
		m.UserId = types.StringNull()
	}
	if dto.Email != nil {
		m.Email = types.StringValue(*dto.Email)
	} else {
		m.Email = types.StringNull()
	}
	return diag.Diagnostics{}
}

func (m userReferenceModel) ToDto(_ context.Context) (sifflet.PublicReferenceByIdOrEmailDto, diag.Diagnostics) {
	var userId *uuid.UUID
	if !m.UserId.IsNull() && !m.UserId.IsUnknown() {
		uid, err := uuid.Parse(m.UserId.ValueString())
		if err != nil {
			return sifflet.PublicReferenceByIdOrEmailDto{}, tfutils.ErrToDiags("Could not parse user ID as UUID", err)
		}
		userId = &uid
	}

	var email *string
	if !m.Email.IsNull() && !m.Email.IsUnknown() {
		e := m.Email.ValueString()
		email = &e
	}

	return sifflet.PublicReferenceByIdOrEmailDto{
		Id:    userId,
		Email: email,
	}, diag.Diagnostics{}
}

package provider

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &userResource{}
	_ resource.ResourceWithConfigure = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *sifflet.ClientWithResponses
}

// Metadata returns the resource type name.
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func UserResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Manage a Sifflet user.",
		MarkdownDescription: "Manage a Sifflet user. See the [Sifflet documentation about access control](https://docs.siffletdata.com/docs/access-control) for more information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "User ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Description: "User email. Also used as a the login identifier. Updates require recreating the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
			"name": schema.StringAttribute{
				Description: "User full name.",
				Required:    true,
			},
			"role": schema.StringAttribute{
				Description: "User system role. Determines a user's access and permissions over Sifflet-level settings. One of 'ADMIN', 'EDITOR', 'VIEWER'.",
				Required:    true,
			},
			"permissions": schema.ListNestedAttribute{
				Description: "Per-domain user permissions. Can not be empty.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain_id": schema.StringAttribute{
							Description: "Domain ID. This can be retrieved from the domain details page from the Sifflet UI (there's no public API for this as of this writing).",
							Required:    true,
						},
						"domain_role": schema.StringAttribute{
							Description: "User role in the domain. One of 'EDITOR', 'VIEWER'.",
							Required:    true,
						},
					},
				},
				// The API will return an error if the list is empty. It's an easy mistake to make, so add client-side validation.
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

type UserModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Email       types.String `tfsdk:"email"`
	Role        types.String `tfsdk:"role"`
	Permissions types.List   `tfsdk:"permissions"`
}

func (m UserModel) GetPermissionsModel() ([]PermissionModel, diag.Diagnostics) {
	permissions := make([]PermissionModel, 0, len(m.Permissions.Elements()))
	diags := m.Permissions.ElementsAs(context.Background(), &permissions, false)
	return permissions, diags
}

func (m UserModel) ToCreateDto(ctx context.Context) (sifflet.PublicUserCreateDto, diag.Diagnostics) {
	permissionsModel, diags := m.GetPermissionsModel()
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

func (m UserModel) ToUpdateDto(ctx context.Context) (sifflet.PublicUserUpdateDto, diag.Diagnostics) {
	permissionsModel, diags := m.GetPermissionsModel()
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

func ToUserModel(ctx context.Context, userDto sifflet.PublicUserGetDto) (UserModel, diag.Diagnostics) {
	newPermissionsModel := make([]PermissionModel, len(userDto.Permissions))
	for i, permission := range userDto.Permissions {
		dto, diags := ToPermissionModel(permission)
		if diags.HasError() {
			return UserModel{}, diags
		}
		newPermissionsModel[i] = dto
	}
	permissionsList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: PermissionModel{}.AttributeTypes()}, newPermissionsModel)
	if diags.HasError() {
		return UserModel{}, diags
	}

	return UserModel{
		Id:          types.StringValue(userDto.Id.String()),
		Name:        types.StringValue(userDto.Name),
		Email:       types.StringValue(userDto.Email),
		Role:        types.StringValue(string(userDto.Role)),
		Permissions: permissionsList,
	}, diag.Diagnostics{}
}

func ToPermissionModel(permissionDto sifflet.PublicUserPermissionAssignmentDto) (PermissionModel, diag.Diagnostics) {
	domainRole := "VIEWER"
	if permissionDto.DomainRole != nil {
		domainRole = string(*permissionDto.DomainRole)
	}
	return PermissionModel{
		DomainId:   types.StringValue(permissionDto.DomainId.String()),
		DomainRole: types.StringValue(domainRole),
	}, diag.Diagnostics{}
}

type PermissionModel struct {
	DomainId   types.String `tfsdk:"domain_id"`
	DomainRole types.String `tfsdk:"domain_role"`
}

func (m PermissionModel) ToDto() (sifflet.PublicUserPermissionAssignmentDto, diag.Diagnostics) {
	uid, err := uuid.Parse(m.DomainId.ValueString())
	if err != nil {
		return sifflet.PublicUserPermissionAssignmentDto{},
			diag.Diagnostics{
				diag.NewErrorDiagnostic("Could not parse domain ID as UUID", err.Error()),
			}
	}
	if err != nil {
		return sifflet.PublicUserPermissionAssignmentDto{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("Invalid domain role", err.Error()),
		}
	}
	role := sifflet.PublicUserPermissionAssignmentDtoDomainRole(m.DomainRole.ValueString())
	return sifflet.PublicUserPermissionAssignmentDto{
		DomainId:   uid,
		DomainRole: &role,
	}, diag.Diagnostics{}
}

func (m PermissionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"domain_id":   types.StringType,
		"domain_role": types.StringType,
	}
}

func (r *userResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = UserResourceSchema(ctx)
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userDto, diags := plan.ToCreateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	userResponse, err := r.client.PublicCreateUserWithResponse(ctx, userDto)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create user", err.Error())
		return
	}

	if userResponse.StatusCode() != http.StatusCreated {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to create user",
			userResponse.StatusCode(), userResponse.Body,
		)
		return
	}

	newState, diags := ToUserModel(ctx, *userResponse.JSON201)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	uid, err := uuid.Parse(id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read user: could not parse user ID as UUID", err.Error())
		return
	}

	userResponse, err := r.client.PublicGetUserWithResponse(ctx, uid)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read user", err.Error())
		return
	}

	if userResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read user",
			userResponse.StatusCode(), userResponse.Body)
		resp.State.RemoveResource(ctx)
		return
	}

	newState, diags := ToUserModel(ctx, *userResponse.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.Id.ValueString()
	uid, err := uuid.Parse(id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update user: could not parse user ID as UUID", err.Error())
		return
	}

	updateDto, diags := plan.ToUpdateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateResponse, err := r.client.PublicUpdateUserWithResponse(ctx, uid, updateDto)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update user", err.Error())
		return
	}

	if updateResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to update user",
			updateResponse.StatusCode(), updateResponse.Body,
		)
		return
	}

	newState, diags := ToUserModel(ctx, *updateResponse.JSON200)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	uid, err := uuid.Parse(id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete user: could not parse user ID as UUID", err.Error())
		return
	}

	userResponse, _ := r.client.PublicDeleteUserWithResponse(ctx, uid)

	if userResponse.StatusCode() != http.StatusNoContent {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to delete user",
			userResponse.StatusCode(), userResponse.Body,
		)
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*apiclients.HttpClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *HttpClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = clients.Client
}

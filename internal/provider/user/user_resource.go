package user

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                 = &userResource{}
	_ resource.ResourceWithConfigure    = &userResource{}
	_ resource.ResourceWithUpgradeState = &userResource{}
)

func newUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *sifflet.ClientWithResponses
}

// Metadata returns the resource type name.
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func userResourceSchema() schema.Schema {
	return schema.Schema{
		Version:             1,
		Description:         "Manage a Sifflet user.",
		MarkdownDescription: "Manage a Sifflet user. See the [Sifflet documentation about access control](https://docs.siffletdata.com/docs/access-control) for more information.\n\nWarning: creating a user will send an email to the specified email address giving them instructions on how to connect to the Sifflet environment.",
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
			"permissions": schema.SetNestedAttribute{
				Description: "Per-domain user permissions. Can not be empty.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain_id": schema.StringAttribute{
							Description: "Domain ID. This can be retrieved from the domain details page from the Sifflet UI or from the sifflet_domain data source or resource.",
							Required:    true,
						},
						"domain_role": schema.StringAttribute{
							Description: "User role in the domain. One of 'EDITOR', 'VIEWER', 'CATALOG_EDITOR', 'MONITOR_RESPONDER'.",
							Required:    true,
						},
					},
				},
				// The API will return an error if the set is empty. It's an easy mistake to make, so add client-side validation.
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"auth_types": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "Authorized authentication types for the user. Possible values are 'SAML2' and 'LOGIN_PASSWORD'. Default is ['SAML2'] if your Sifflet instance has SSO enabled, and ['LOGIN_PASSWORD', 'SAML2'] otherwise.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = userResourceSchema()
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, cancel := tfutils.WithDefaultCreateTimeout(ctx)
	defer cancel()

	var plan userModel
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

	var newState userModel
	diags = newState.FromDto(ctx, *userResponse.JSON201)
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
	ctx, cancel := tfutils.WithDefaultReadTimeout(ctx)
	defer cancel()

	var state userModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, diags := state.ModelId()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userResponse, err := r.client.PublicGetUserWithResponse(ctx, id)
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

	var newState userModel
	diags = newState.FromDto(ctx, *userResponse.JSON200)
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

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, cancel := tfutils.WithDefaultUpdateTimeout(ctx)
	defer cancel()

	var plan userModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, diags := plan.ModelId()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateDto, diags := plan.ToUpdateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateResponse, err := r.client.PublicUpdateUserWithResponse(ctx, id, updateDto)
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

	var newState userModel
	diags = newState.FromDto(ctx, *updateResponse.JSON200)
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
	ctx, cancel := tfutils.WithDefaultDeleteTimeout(ctx)
	defer cancel()

	var state userModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, diags := state.ModelId()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userResponse, _ := r.client.PublicDeleteUserWithResponse(ctx, id)

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

func (r *userResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	// Build v0 schema by copying v1 and overriding only the changed attribute
	v0Schema := userResourceSchema()
	v0Schema.Version = 0

	// Override permissions to be ListNestedAttribute (v0) instead of SetNestedAttribute (v1)
	v0Schema.Attributes["permissions"] = schema.ListNestedAttribute{
		Description: "Per-domain user permissions. Can not be empty.",
		Required:    true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"domain_id": schema.StringAttribute{
					Description: "Domain ID.",
					Required:    true,
				},
				"domain_role": schema.StringAttribute{
					Description: "User role in the domain.",
					Required:    true,
				},
			},
		},
	}

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &v0Schema,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				type userModelV0 struct {
					Id          types.String `tfsdk:"id"`
					Name        types.String `tfsdk:"name"`
					Email       types.String `tfsdk:"email"`
					Role        types.String `tfsdk:"role"`
					Permissions types.List   `tfsdk:"permissions"`
					AuthTypes   types.Set    `tfsdk:"auth_types"`
				}

				var priorState userModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// Convert permissions from List to Set
				permissionAttrTypes := permissionModel{}.AttributeTypes()
				var permissionsUpgraded types.Set
				if !priorState.Permissions.IsNull() && !priorState.Permissions.IsUnknown() {
					var permissionObjs []types.Object
					resp.Diagnostics.Append(priorState.Permissions.ElementsAs(ctx, &permissionObjs, false)...)
					if resp.Diagnostics.HasError() {
						return
					}
					var ds basetypes.SetValue
					ds, resp.Diagnostics = types.SetValueFrom(ctx, types.ObjectType{AttrTypes: permissionAttrTypes}, permissionObjs)
					if resp.Diagnostics.HasError() {
						return
					}
					permissionsUpgraded = ds
				} else {
					permissionsUpgraded = types.SetNull(types.ObjectType{AttrTypes: permissionAttrTypes})
				}

				upgradedState := userModel{
					Id:          priorState.Id,
					Name:        priorState.Name,
					Email:       priorState.Email,
					Role:        priorState.Role,
					Permissions: permissionsUpgraded,
					AuthTypes:   priorState.AuthTypes,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedState)...)
			},
		},
	}
}

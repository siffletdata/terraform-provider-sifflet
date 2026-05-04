package team

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ resource.Resource              = &teamResource{}
	_ resource.ResourceWithConfigure = &teamResource{}
)

func newTeamResource() resource.Resource {
	return &teamResource{}
}

type teamResource struct {
	client *sifflet.ClientWithResponses
}

// Metadata returns the resource type name.
func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manage a Sifflet team.",
		MarkdownDescription: "Manage a Sifflet team. See the [Sifflet documentation about teams](https://docs.siffletdata.com/docs/teams) for more information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Team ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Team name.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Team description.",
				Optional:    true,
			},
			"domain_permissions": schema.SetNestedAttribute{
				Description: "Domain permissions granted to the team.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain_id": schema.StringAttribute{
							Description: "Domain ID. This can be retrieved from the domain details page from the Sifflet UI or from the sifflet_domain data source or resource.",
							Required:    true,
						},
						"domain_role": schema.StringAttribute{
							Description: "Team role in the domain. One of 'EDITOR', 'VIEWER', 'CATALOG_EDITOR', 'MONITOR_RESPONDER'.",
							Required:    true,
						},
					},
				},
			},
			"users": schema.SetNestedAttribute{
				Description: "Users belonging to the team. Each user must be specified by either user_id or email.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.StringAttribute{
							Description: "User ID. Either user_id or email must be specified.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("email"),
								),
							},
						},
						"email": schema.StringAttribute{
							Description: "User email. Either user_id or email must be specified.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("user_id"),
								),
							},
						},
					},
				},
			},
		},
	}
}

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, cancel := tfutils.WithDefaultCreateTimeout(ctx)
	defer cancel()

	var plan teamModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamDto, diags := plan.ToCreateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	teamResponse, err := r.client.PublicCreateTeamWithResponse(ctx, teamDto)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create team", err.Error())
		return
	}

	if teamResponse.StatusCode() != http.StatusCreated {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to create team",
			teamResponse.StatusCode(), teamResponse.Body,
		)
		return
	}

	var newState teamModel
	diags = newState.FromDto(ctx, *teamResponse.JSON201)
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

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, cancel := tfutils.WithDefaultReadTimeout(ctx)
	defer cancel()

	var state teamModel
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

	teamResponse, err := r.client.PublicGetTeamWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read team", err.Error())
		return
	}

	if teamResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read team",
			teamResponse.StatusCode(), teamResponse.Body)
		resp.State.RemoveResource(ctx)
		return
	}

	var newState teamModel
	diags = newState.FromDto(ctx, *teamResponse.JSON200)
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

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, cancel := tfutils.WithDefaultUpdateTimeout(ctx)
	defer cancel()

	var plan teamModel
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

	updateResponse, err := r.client.PublicUpdateTeamWithResponse(ctx, id, updateDto)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update team", err.Error())
		return
	}

	if updateResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to update team",
			updateResponse.StatusCode(), updateResponse.Body,
		)
		return
	}

	var newState teamModel
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

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx, cancel := tfutils.WithDefaultDeleteTimeout(ctx)
	defer cancel()

	var state teamModel
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

	teamResponse, _ := r.client.PublicDeleteTeamWithResponse(ctx, id)

	if teamResponse.StatusCode() != http.StatusNoContent {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to delete team",
			teamResponse.StatusCode(), teamResponse.Body,
		)
		return
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

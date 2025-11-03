package domain

import (
	"context"
	"fmt"
	"net/http"

	"terraform-provider-sifflet/internal/apiclients"
	sifflet "terraform-provider-sifflet/internal/client"
	"terraform-provider-sifflet/internal/tfutils"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                     = &domainResource{}
	_ resource.ResourceWithConfigure        = &domainResource{}
	_ resource.ResourceWithConfigValidators = &domainResource{}
)

func newDomainResource() resource.Resource {
	return &domainResource{}
}

type domainResource struct {
	client *sifflet.ClientWithResponses
}

// Metadata returns the resource type name.
func (r *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func domainResourceSchema() schema.Schema {
	return schema.Schema{
		Description:         "A Sifflet domain.",
		MarkdownDescription: "A Sifflet domain. A domain represents a subset of assets. Domains are used to provide a view of specific business areas, and provide different access to different teams.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the domain.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the domain.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the domain.",
				Optional:    true,
			},
			"dynamic_content_definition": schema.SingleNestedAttribute{
				Description: "The dynamic content definition of the domain. At least one of dynamic_content_definition or static_content_definition must be provided.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"logical_operator": schema.StringAttribute{
						Description: "The logical operator to use between conditions. One of 'AND' or 'OR'.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("AND", "OR"),
						},
					},
					"conditions": schema.ListNestedAttribute{
						Description: "The conditions to use to define the dynamic content of the domain.",
						Required:    true,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"logical_operator": schema.StringAttribute{
									Description: "The logical operator for this condition. One of 'IS' or 'IS_NOT'.",
									Required:    true,
									Validators: []validator.String{
										stringvalidator.OneOf("IS", "IS_NOT"),
									},
								},
								"schema_uris": schema.ListAttribute{
									Description: "The source schemas to filter assets by in the dynamic condition, in URI format. More about URIs here: https://docs.siffletdata.com/docs/uris.",
									Optional:    true,
									ElementType: types.StringType,
								},
								"tags": schema.ListNestedAttribute{
									Description: "The tags to filter assets by in the dynamic condition.",
									Optional:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												Description: "The ID of the tag. If provided, name and kind must be omitted.",
												Optional:    true,
												Computed:    true,
											},
											"name": schema.StringAttribute{
												Description: "The name of the tag. If provided, id must be omitted.",
												Optional:    true,
												Computed:    true,
												Validators: []validator.String{
													stringvalidator.ExactlyOneOf(
														path.MatchRelative().AtParent().AtName("id"),
														path.MatchRelative(),
													),
												},
											},
											"kind": schema.StringAttribute{
												Description: "The kind of the tag. If provided, name must be provided. Use this field for disambiguation when the tag name matches tags from multiple source types. One of 'BIGQUERY_EXTERNAL', 'DATABRICKS_EXTERNAL', 'DBT_EXTERNAL', 'SNOWFLAKE_EXTERNAL'. ",
												Optional:    true,
												Computed:    true,
												Validators: []validator.String{
													stringvalidator.OneOf("BIGQUERY_EXTERNAL", "DATABRICKS_EXTERNAL", "DBT_EXTERNAL", "SNOWFLAKE_EXTERNAL"),
													stringvalidator.ConflictsWith(
														path.MatchRelative().AtParent().AtName("id"),
													),
												},
											},
										},
									},
									Validators: []validator.List{
										listvalidator.ExactlyOneOf(
											path.MatchRelative().AtParent().AtName("schema_uris"),
											path.MatchRelative(),
										),
									},
								},
							},
						},
					},
				},
			},
			"static_content_definition": schema.SingleNestedAttribute{
				Description: "The static content definition of the domain. At least one of dynamic_content_definition or static_content_definition must be provided.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"asset_uris": schema.SetAttribute{
						ElementType: types.StringType,
						Description: "The URIs of the assets to include in the domain.",
						Required:    true,
						Validators: []validator.Set{
							setvalidator.SizeAtLeast(1),
						},
					},
				},
			},
		},
	}
}

// We want to enforce that one of dynamic_content_definition or static_content_definition is provided.
func (r domainResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("dynamic_content_definition"),
			path.MatchRoot("static_content_definition"),
		),
	}
}

// Schema defines the schema for the resource.
func (r *domainResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = domainResourceSchema()
}

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, cancel := tfutils.WithDefaultReadTimeout(ctx)
	defer cancel()

	var plan domainModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainDto, diags := plan.ToCreateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainResponse, err := r.client.PublicCreateDomainWithResponse(ctx, domainDto)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create domain", err.Error())
		return
	}

	if domainResponse.StatusCode() != http.StatusCreated {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to create domain", domainResponse.StatusCode(), domainResponse.Body,
		)
		return
	}

	var newState domainModel
	diags = newState.FromDto(ctx, *domainResponse.JSON201)
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

func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, cancel := tfutils.WithDefaultReadTimeout(ctx)
	defer cancel()

	var state domainModel
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

	domainResponse, err := r.client.PublicGetDomainWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read domain", err.Error())
		return
	}

	if domainResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read domain", domainResponse.StatusCode(), domainResponse.Body,
		)
		return
	}

	var newState domainModel
	diags = newState.FromDto(ctx, *domainResponse.JSON200)
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

func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, cancel := tfutils.WithDefaultUpdateTimeout(ctx)
	defer cancel()

	var plan domainModel
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

	domainDto, diags := plan.ToUpdateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainResponse, err := r.client.PublicUpdateDomainWithResponse(ctx, id, domainDto)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update domain", err.Error())
		return
	}

	if domainResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to update domain", domainResponse.StatusCode(), domainResponse.Body,
		)
		return
	}

	var newState domainModel
	diags = newState.FromDto(ctx, *domainResponse.JSON200)
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

func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx, cancel := tfutils.WithDefaultDeleteTimeout(ctx)
	defer cancel()

	var state domainModel
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

	domainResponse, err := r.client.PublicDeleteDomainWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete domain", err.Error())
		return
	}

	if domainResponse.StatusCode() != http.StatusNoContent {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to delete domain", domainResponse.StatusCode(), domainResponse.Body,
		)
		return
	}
}

func (r *domainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *domainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

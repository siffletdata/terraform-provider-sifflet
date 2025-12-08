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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ resource.Resource                     = &domainResource{}
	_ resource.ResourceWithConfigure        = &domainResource{}
	_ resource.ResourceWithConfigValidators = &domainResource{}
	_ resource.ResourceWithUpgradeState     = &domainResource{}
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
		Version:             1,
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
								"schema_uris": schema.SetAttribute{
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

// upgradeDynamicContentDefinitionV0ToV1 converts schema_uris from List to Set in dynamic content definition conditions.
func upgradeDynamicContentDefinitionV0ToV1(ctx context.Context, dynamicV0Obj types.Object) (types.Object, diag.Diagnostics) {
	type dynamicContentDefinitionConditionModelV0 struct {
		LogicalOperator types.String `tfsdk:"logical_operator"`
		SchemaUris      types.List   `tfsdk:"schema_uris"`
		Tags            types.List   `tfsdk:"tags"`
	}

	type dynamicContentDefinitionModelV0 struct {
		LogicalOperator types.String `tfsdk:"logical_operator"`
		Conditions      types.List   `tfsdk:"conditions"`
	}

	var dynamicV0 dynamicContentDefinitionModelV0
	diags := dynamicV0Obj.As(ctx, &dynamicV0, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return types.Object{}, diags
	}

	var conditionsV0 []types.Object
	diags.Append(dynamicV0.Conditions.ElementsAs(ctx, &conditionsV0, false)...)
	if diags.HasError() {
		return types.Object{}, diags
	}

	upgradeCondition := func(conditionObjV0 types.Object) (types.Object, diag.Diagnostics) {
		var conditionV0 dynamicContentDefinitionConditionModelV0
		diags := conditionObjV0.As(ctx, &conditionV0, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return types.Object{}, diags
		}

		// Convert schema_uris from List to Set
		var schemaUrisUpgraded types.Set
		if !conditionV0.SchemaUris.IsNull() && !conditionV0.SchemaUris.IsUnknown() {
			var schemaUrisList []types.String
			diags.Append(conditionV0.SchemaUris.ElementsAs(ctx, &schemaUrisList, false)...)
			if diags.HasError() {
				return types.Object{}, diags
			}
			schemaUrisUpgraded, diags = types.SetValueFrom(ctx, types.StringType, schemaUrisList)
			if diags.HasError() {
				return types.Object{}, diags
			}
		} else {
			schemaUrisUpgraded = types.SetNull(types.StringType)
		}

		upgradedCondition := dynamicContentDefinitionConditionModel{
			LogicalOperator: conditionV0.LogicalOperator,
			SchemaUris:      schemaUrisUpgraded,
			Tags:            conditionV0.Tags,
		}

		return types.ObjectValueFrom(ctx, upgradedCondition.AttributeTypes(), upgradedCondition)
	}

	upgradedConditions, diags := tfutils.MapWithDiagnostics(conditionsV0, upgradeCondition)
	if diags.HasError() {
		return types.Object{}, diags
	}

	upgradedConditionsList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: dynamicContentDefinitionConditionModel{}.AttributeTypes()}, upgradedConditions)
	if diags.HasError() {
		return types.Object{}, diags
	}

	upgradedDynamic := dynamicContentDefinitionModel{
		LogicalOperator: dynamicV0.LogicalOperator,
		Conditions:      upgradedConditionsList,
	}

	return types.ObjectValueFrom(ctx, upgradedDynamic.AttributeTypes(), upgradedDynamic)
}

func (r *domainResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	// Build v0 schema by copying v1 and overriding only the changed attribute
	v0Schema := domainResourceSchema()
	v0Schema.Version = 0

	// Override schema_uris to be ListAttribute (v0) instead of SetAttribute (v1)
	dynamicContentDef, ok := v0Schema.Attributes["dynamic_content_definition"].(schema.SingleNestedAttribute)
	if !ok {
		panic("dynamic_content_definition is not a SingleNestedAttribute in the current domain resource schema (this is a bug in the provider)")
	}
	conditions, ok := dynamicContentDef.Attributes["conditions"].(schema.ListNestedAttribute)
	if !ok {
		panic("conditions is not a ListNestedAttribute in the current domain resource schema (this is a bug in the provider)")
	}
	conditions.NestedObject.Attributes["schema_uris"] = schema.ListAttribute{
		Description: "The source schemas to filter assets by in the dynamic condition, in URI format.",
		Optional:    true,
		ElementType: types.StringType,
	}
	dynamicContentDef.Attributes["conditions"] = conditions
	v0Schema.Attributes["dynamic_content_definition"] = dynamicContentDef

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &v0Schema,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				type domainModelV0 struct {
					Id                       types.String `tfsdk:"id"`
					Name                     types.String `tfsdk:"name"`
					Description              types.String `tfsdk:"description"`
					DynamicContentDefinition types.Object `tfsdk:"dynamic_content_definition"`
					StaticContentDefinition  types.Object `tfsdk:"static_content_definition"`
				}

				var priorState domainModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgradedState := domainModel{
					Id:                      priorState.Id,
					Name:                    priorState.Name,
					Description:             priorState.Description,
					StaticContentDefinition: priorState.StaticContentDefinition,
				}

				if !priorState.DynamicContentDefinition.IsNull() && !priorState.DynamicContentDefinition.IsUnknown() {
					upgradedDynamicObj, diags := upgradeDynamicContentDefinitionV0ToV1(ctx, priorState.DynamicContentDefinition)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
					upgradedState.DynamicContentDefinition = upgradedDynamicObj
				} else {
					upgradedState.DynamicContentDefinition = types.ObjectNull(dynamicContentDefinitionModel{}.AttributeTypes())
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedState)...)
			},
		},
	}
}

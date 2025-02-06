package tag

import (
	"context"
	"fmt"
	"net/http"

	sifflet "terraform-provider-sifflet/internal/alphaclient"
	"terraform-provider-sifflet/internal/apiclients"
	"terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ resource.Resource              = &tagResource{}
	_ resource.ResourceWithConfigure = &tagResource{}
)

func newTagResource() resource.Resource {
	return &tagResource{}
}

type tagResource struct {
	client *sifflet.ClientWithResponses
}

// Metadata returns the resource type name.
func (r *tagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

// Schema defines the schema for the resource.
func (r *tagResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = tagResourceSchema()
}

func tagResourceSchema() schema.Schema {
	return schema.Schema{
		Version:     1,
		Description: "Manage a Sifflet tag.",
		MarkdownDescription: `Tags are used to classify data in Sifflet.

This resource manages 'regular' tags. See the [Sifflet documentation](https://docs.siffletdata.com/docs/tags) for more about tag types.",
		`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Tag ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Tag name.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "Tag description.",
				Optional:    true,
			},
		},
	}
}

func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagDto, diags := plan.ToCreateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagResponse, _ := r.client.CreateTagWithResponse(ctx, tagDto)

	if tagResponse.StatusCode() != http.StatusCreated {
		client.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to create tag", tagResponse.StatusCode(), tagResponse.Body)
		resp.State.RemoveResource(ctx)
		return
	}

	var newState tagModel
	diags = newState.FromDto(ctx, *tagResponse.JSON201)
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

func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tagModel
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

	tagResponse, err := r.client.GetTagByIdWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read tag", err.Error())
		return
	}

	if tagResponse.StatusCode() != http.StatusOK {
		client.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read tag", tagResponse.StatusCode(), tagResponse.Body)
		resp.State.RemoveResource(ctx)
		return
	}

	var newState tagModel
	diags = newState.FromDto(ctx, *tagResponse.JSON200)
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

func (r *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tagModel
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

	tagDto, diags := plan.ToUpdateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagResponse, err := r.client.UpdateTagWithResponse(ctx, id, tagDto)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update tag", err.Error())
		return
	}

	if tagResponse.StatusCode() != http.StatusOK {
		client.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to update tag", tagResponse.StatusCode(), tagResponse.Body)
		resp.State.RemoveResource(ctx)
		return
	}

	var newState tagModel
	diags = newState.FromDto(ctx, *tagResponse.JSON200)
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

func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tagModel
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

	tagResponse, err := r.client.DeleteTagWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to delete tag", err.Error())
		return
	}

	if tagResponse.StatusCode() != http.StatusNoContent {
		client.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to delete tag", tagResponse.StatusCode(), tagResponse.Body)
		return
	}
}

func (r *tagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *tagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = clients.AlphaClient
}

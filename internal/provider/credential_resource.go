package provider

import (
	"context"
	"fmt"
	"net/http"

	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var (
	_ resource.Resource              = &credentialResource{}
	_ resource.ResourceWithConfigure = &credentialResource{}
)

func NewCredentialResource() resource.Resource {
	return &credentialResource{}
}

type credentialResource struct {
	client *sifflet.ClientWithResponses
}

// Metadata returns the resource type name.
func (r *credentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

func CredentialResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		// TODO docs
		Description: "A credential resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the credential",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the credential",
				Optional:    true,
			},
			"value": schema.StringAttribute{
				Description: "The value of the credential",
				Sensitive:   true,
			},
		},
	}

}

type CredentialDto struct {
	Name        string  `tfsdk:"name"`
	Description *string `tfsdk:"description"`
	Value       string  `tfsdk:"value"`
}

func (r *credentialResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CredentialResourceSchema(ctx)
}

func (r *credentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CredentialDto // FIXME should this be named "state"?
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialDto := sifflet.PublicCredentialCreateDto{
		Name:        plan.Name,
		Description: plan.Description,
		Value:       plan.Value,
	}

	credentialResponse, err := r.client.PublicCreateCredentialWithResponse(ctx, credentialDto)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create credential", err.Error())
		return
	}

	if credentialResponse.StatusCode() != http.StatusCreated {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to create credential",
			credentialResponse.StatusCode(), credentialResponse.Body,
		)
		resp.State.RemoveResource(ctx) // FIXME: is it necessary?
		return
	}

	plan.Name = credentialDto.Name
	plan.Description = credentialDto.Description
	plan.Value = credentialDto.Value

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// FIXME/ log all API responses (can probably be done at the client level instead of each function)

func (r *credentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CredentialDto
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Name

	credentialResponse, err := r.client.PublicGetCredentialWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read credential",
			err.Error(),
		)
		return
	}

	if credentialResponse.StatusCode() == http.StatusNotFound {
		// FIXME is that the correct logic?
		resp.State.RemoveResource(ctx)
		return
	}

	if credentialResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read credential",
			credentialResponse.StatusCode(), credentialResponse.Body)
		resp.State.RemoveResource(ctx)
		return
	}

	state = CredentialDto{
		Name:        credentialResponse.JSON200.Name,
		Description: credentialResponse.JSON200.Description,
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *credentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO
}

func (r *credentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CredentialDto
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Name

	credentialResponse, _ := r.client.PublicDeleteCredentialWithResponse(ctx, id)

	if credentialResponse.StatusCode() != http.StatusNoContent {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to delete credential",
			credentialResponse.StatusCode(), credentialResponse.Body,
		)
		resp.State.RemoveResource(ctx)
		return
	}

}

func (r *credentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *credentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*httpClients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *httpClients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = clients.Client
}

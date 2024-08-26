package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	sifflet "terraform-provider-sifflet/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
		Description:         "A credential resource.",
		MarkdownDescription: "Credentials are used to store secret source connection information, such as username, passwords, service account keys, or API tokens",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				// TODO add validation (https://developer.hashicorp.com/terraform/plugin/framework/validation#attribute-validation)
				Description: "The name of the credential. Must only contain alphanumeric characters. Must be uniquein the Sifflet instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the credential.",
				Optional:    true,
			},
			"value": schema.StringAttribute{
				Description: "The value of the credential. Due to API limitations, Terraform can't detect changes to this value made outside of Terraform. Mandatory when asking Terraform to create the resource; otherwise, if the resource is imported or was created during a previous apply, this value is optional.",
				Sensitive:   true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}

}

type CredentialDto struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Value       types.String `tfsdk:"value"`
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

	if plan.Value.IsNull() {
		resp.Diagnostics.AddError("Value is required", "The value attribute is required when creating a credential.")
		return
	}

	credentialDto := sifflet.PublicCredentialCreateDto{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueStringPointer(),
		Value:       plan.Value.ValueString(),
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

	plan.Name = types.StringValue(credentialDto.Name)
	plan.Description = types.StringPointerValue(credentialDto.Description)
	plan.Value = types.StringValue(credentialDto.Value)

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

	id := state.Name.ValueString()

	maxAttempts := 20
	var credentialResponse *sifflet.PublicGetCredentialResponse
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		credentialResponse, err = r.client.PublicGetCredentialWithResponse(ctx, id)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read credential",
				err.Error(),
			)
			return
		}

		if credentialResponse.StatusCode() == http.StatusNotFound {
			// Retry a few times, as there's a delay in the API (eventual consistency)
			if attempt < maxAttempts {
				time.Sleep(200 * time.Millisecond)
				continue
			}
		} else {
			break
		}
	}

	if credentialResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read credential",
			credentialResponse.StatusCode(), credentialResponse.Body)
		resp.State.RemoveResource(ctx)
		return
	}

	state = CredentialDto{
		Name:        types.StringValue(credentialResponse.JSON200.Name),
		Description: types.StringPointerValue(credentialResponse.JSON200.Description),
		// TODO: The API doesn't include any way to detect if the secret value has changed (like a version field).
		// See PLTE-901.
		// In the meantime, let's copy the previous value from the state, if any. This won't allow Terraform to detect whether the value has changed outside of Terraform though.
		Value: state.Value,
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *credentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CredentialDto
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.Name.ValueString()
	body := sifflet.PublicUpdateCredentialJSONRequestBody{
		Description: plan.Description.ValueStringPointer(),
		Value:       plan.Value.ValueStringPointer(),
	}

	updateResponse, err := r.client.PublicUpdateCredentialWithResponse(ctx, id, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update credential", err.Error())
		return
	}

	if updateResponse.StatusCode() != http.StatusNoContent {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to update credential",
			updateResponse.StatusCode(), updateResponse.Body,
		)
		return
	}

	// Since the credential API is eventually consistent, we wait until we read back the description that we wrote.
	// Otherwise, the next read by Terraform might return the old value, which would generate an error (inconsistent plan).
	// Reading the credential description is currently the only way we can know whether the credential was updated,  the API doesn't include any way to detect if the secret value has changed (like a version field).
	// See PLTE-901.
	maxAttempts := 20
	var credentialResponse *sifflet.PublicGetCredentialResponse
	for attempt := 0; attempt < maxAttempts; attempt++ {
		credentialResponse, err = r.client.PublicGetCredentialWithResponse(ctx, id)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read back credential after updating it",
				err.Error(),
			)
			return
		}

		if credentialResponse.StatusCode() != http.StatusOK {
			sifflet.HandleHttpErrorAsProblem(
				ctx, &resp.Diagnostics, "Unable to read credential after updating it",
				credentialResponse.StatusCode(), credentialResponse.Body)
			resp.State.RemoveResource(ctx)
			return
		}

		if credentialResponse.JSON200.Description == body.Description {
			break
		}
		time.Sleep(200 * time.Millisecond)
		// If we exhausted the attempts, try to proceed anyway. We still hope that when Terraform reads
		// back the value, it will be updated by that time.
	}

	plan.Description = types.StringPointerValue(body.Description)
	plan.Value = types.StringPointerValue(body.Value)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *credentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CredentialDto
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Name.ValueString()

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
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
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

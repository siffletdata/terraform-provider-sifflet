package credentials

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &credentialsResource{}
	_ resource.ResourceWithConfigure = &credentialsResource{}
)

func newCredentialResource() resource.Resource {
	return &credentialsResource{}
}

type credentialsResource struct {
	client *sifflet.ClientWithResponses
}

// Metadata returns the resource type name.
func (r *credentialsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credentials"
}

func CredentialResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "A credentials resource.",
		MarkdownDescription: "Credentials are used to store secret source connection information, such as usernames, passwords, service account keys, or API tokens.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the credentials. Must start and end with a letter, and contain only letters, digits and hyphens. Must be unique in the Sifflet instance.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*[a-zA-Z]$`),
						"must start and end with a letter, and contain only letters, digits, and hyphens",
					),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the credentials.",
				Optional:    true,
			},
			"value": schema.StringAttribute{
				Description: "The value of the credentials. Due to API limitations, Terraform can't detect changes to this value made outside of Terraform. Mandatory when asking Terraform to create the resource; otherwise, if the resource is imported or was created during a previous apply, this value is optional.",
				Sensitive:   true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}

}

func (r *credentialsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CredentialResourceSchema(ctx)
}

func (r *credentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, cancel := tfutils.WithDefaultCreateTimeout(ctx)
	defer cancel()

	var plan credentialModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Value.IsNull() {
		resp.Diagnostics.AddError("Value is required", "The value attribute is required when creating credentials.")
		return
	}

	credentialsDto, diags := plan.ToCreateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialsResponse, err := r.client.PublicCreateCredentialsWithResponse(ctx, credentialsDto)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create credentials", err.Error())
		return
	}

	if credentialsResponse.StatusCode() != http.StatusCreated {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to create credentials",
			credentialsResponse.StatusCode(), credentialsResponse.Body,
		)
		return
	}

	plan.Name = types.StringValue(credentialsDto.Name)
	plan.Description = types.StringPointerValue(credentialsDto.Description)
	plan.Value = types.StringValue(credentialsDto.Value)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Since the credentials API is eventually consistent, we wait until we can read back the credentials that we created.
	// Otherwise, further operations with these credentials (such as "create a datasource referencing these credentials") might fail.
	maxAttempts := 20
	for attempt := 0; attempt < maxAttempts; attempt++ {
		_, err = r.client.PublicGetCredentialsWithResponse(ctx, credentialsDto.Name)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read back credentials after creating them", err.Error())
			return
		}

		if credentialsResponse.StatusCode() == http.StatusOK {
			break
		}

		time.Sleep(200 * time.Millisecond)
		// If we exhausted the attempts, try to proceed anyway. We still hope that when Terraform reads
		// back the value, it will be updated by that time.
	}

}

func (r *credentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx, cancel := tfutils.WithDefaultReadTimeout(ctx)
	defer cancel()

	var state credentialModel
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

	maxAttempts := 20
	var err error
	var credentialsResponse *sifflet.PublicGetCredentialsResponse
	for attempt := 0; attempt < maxAttempts; attempt++ {
		credentialsResponse, err = r.client.PublicGetCredentialsWithResponse(ctx, id)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read credential",
				err.Error(),
			)
			return
		}

		if credentialsResponse.StatusCode() == http.StatusNotFound {
			// Retry a few times, as there's a delay in the API (eventual consistency)
			if attempt < maxAttempts {
				time.Sleep(200 * time.Millisecond)
				continue
			}
		} else {
			break
		}
	}

	if credentialsResponse.StatusCode() != http.StatusOK {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to read credentials",
			credentialsResponse.StatusCode(), credentialsResponse.Body)
		return
	}

	state = credentialModel{
		Name:        types.StringValue(credentialsResponse.JSON200.Name),
		Description: types.StringPointerValue(credentialsResponse.JSON200.Description),
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

func (r *credentialsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, cancel := tfutils.WithDefaultUpdateTimeout(ctx)
	defer cancel()

	var plan credentialModel
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

	body, diags := plan.ToUpdateDto(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateResponse, err := r.client.PublicUpdateCredentialsWithResponse(ctx, id, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update credentials", err.Error())
		return
	}

	if updateResponse.StatusCode() != http.StatusNoContent {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to update credentials",
			updateResponse.StatusCode(), updateResponse.Body,
		)
		return
	}

	// Since the credentials API is eventually consistent, we wait until we read back the description that we wrote.
	// Otherwise, the next read by Terraform might return the old value, which would generate an error (inconsistent plan).
	// Reading the credential description is currently the only way we can know whether the credential was updated,  the API doesn't include any way to detect if the secret value has changed (like a version field).
	// See PLTE-901.
	maxAttempts := 30
	var credentialsResponse *sifflet.PublicGetCredentialsResponse
	for attempt := 0; attempt < maxAttempts; attempt++ {
		credentialsResponse, err = r.client.PublicGetCredentialsWithResponse(ctx, id)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read back credentials after updating it",
				err.Error(),
			)
			return
		}

		if credentialsResponse.StatusCode() != http.StatusOK {
			sifflet.HandleHttpErrorAsProblem(
				ctx, &resp.Diagnostics, "Unable to read credentials after updating it",
				credentialsResponse.StatusCode(), credentialsResponse.Body)
			resp.State.RemoveResource(ctx)
			return
		}

		if credentialsResponse.JSON200.Description == body.Description {
			break
		}
		time.Sleep(300 * time.Millisecond)
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

func (r *credentialsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx, cancel := tfutils.WithDefaultDeleteTimeout(ctx)
	defer cancel()

	var state credentialModel
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

	credentialsResponse, _ := r.client.PublicDeleteCredentialsWithResponse(ctx, id)

	if credentialsResponse.StatusCode() != http.StatusNoContent {
		sifflet.HandleHttpErrorAsProblem(
			ctx, &resp.Diagnostics, "Unable to delete credentials",
			credentialsResponse.StatusCode(), credentialsResponse.Body,
		)
		resp.State.RemoveResource(ctx)
		return
	}

}

func (r *credentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *credentialsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

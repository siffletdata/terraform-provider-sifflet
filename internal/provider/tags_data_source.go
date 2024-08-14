package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	sifflet "terraform-provider-sifflet/internal/alphaclient"
	tag_struct "terraform-provider-sifflet/internal/tag_datasource"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &tagDataSource{}
	_ datasource.DataSourceWithConfigure = &tagDataSource{}
)

func NewTagDataSource() datasource.DataSource {
	return &tagDataSource{}
}

type tagDataSource struct {
	client *sifflet.Client
}

func (d *tagDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*sifflet.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sifflet.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *tagDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}

func (d *tagDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = tag_struct.TagDataSourceSchema(ctx)
}

func (d *tagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tag_struct.SearchCollectionTagDto

	ItemsPerPage := int32(-1)

	params := sifflet.GetAllTagParams{
		ItemsPerPage: &ItemsPerPage,
	}

	itemResponse, err := d.client.GetAllTag(ctx, &params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read data sources",
			err.Error(),
		)
		return
	}
	resBody, _ := io.ReadAll(itemResponse.Body)
	tflog.Debug(ctx, "Response:  "+string(resBody))

	if itemResponse.StatusCode != http.StatusOK {

		var message tag_struct.ErrorMessage
		if err := json.Unmarshal(resBody, &message); err != nil { // Parse []byte to go struct pointer
			resp.Diagnostics.AddError(
				"Can not unmarshal JSON",
				err.Error(),
			)
			return
		}
		resp.Diagnostics.AddError(
			message.Title,
			message.Detail,
		)
		resp.State.RemoveResource(ctx)
		return
	}

	var result sifflet.SearchCollectionTagDto
	if err := json.Unmarshal(resBody, &result); err != nil { // Parse []byte to go struct pointer
		resp.Diagnostics.AddError(
			"Can not unmarshal JSON",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("TotalElements: %d\n", *result.TotalElements))

	state.TotalElements = result.TotalElements

	if state.Data == nil {
		state.Data = &[]tag_struct.TagDto{}
	}

	for _, data := range *result.Data {

		idString := data.Id.String()
		yType := tag_struct.TagDtoType(data.Type)

		var lmd basetypes.StringValue
		if data.LastModifiedDate == nil {
			lmd = types.StringValue("")
		} else {
			lmd = types.StringValue(strconv.FormatInt(*data.LastModifiedDate, 10))
		}

		var cd basetypes.StringValue
		if data.CreatedDate == nil {
			cd = types.StringValue("")
		} else {
			cd = types.StringValue(strconv.FormatInt(*data.CreatedDate, 10))
		}

		var mb basetypes.StringValue
		if data.ModifiedBy == nil {
			mb = types.StringValue("")
		} else {
			mb = types.StringValue(*data.ModifiedBy)
		}

		data_source_catalog_asset := tag_struct.TagDto{
			CreatedBy:        types.StringValue(*data.CreatedBy),
			CreatedDate:      cd,
			Description:      data.Description,
			Editable:         types.BoolValue(*data.Editable),
			LastModifiedDate: lmd,
			ModifiedBy:       mb,
			Name:             &data.Name,
			Type:             &yType,
			Id:               types.StringValue(idString),
		}

		*state.Data = append(*state.Data, data_source_catalog_asset)

	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MachineDataSource{}

func NewMachineDataSource() datasource.DataSource {
	return &MachineDataSource{}
}

// MachineDataSource defines the data source implementation.
type MachineDataSource struct {
	client *client.Client
}

// MachineDataSourceModel describes the data source data model.
type MachineDataSourceModel struct {
	UUID types.String `tfsdk:"uuid"`

	ID     types.Int64  `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
}

func (d *MachineDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine"
}

func (d *MachineDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Machine data source",

		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the machine",
				Optional:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "Machine Melt ID",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Machine Name",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Machine Status",
				Computed:            true,
			},
		},
	}
}

func (d *MachineDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *MachineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MachineDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.Machine().Get(ctx, 17)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read machine, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.Machine.ID)
	data.Name = types.StringValue(result.Machine.Name)
	data.Status = types.StringValue(string(result.Machine.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

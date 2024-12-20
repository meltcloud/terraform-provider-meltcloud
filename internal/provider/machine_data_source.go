package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"strings"
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

	ID            types.Int64            `tfsdk:"id"`
	Name          types.String           `tfsdk:"name"`
	MachinePoolID types.Int64            `tfsdk:"machine_pool_id"`
	Status        types.String           `tfsdk:"status"`
	Labels        []LabelDataSourceModel `tfsdk:"labels"`
}

type LabelDataSourceModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (d *MachineDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine"
}

func (d *MachineDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: machineDesc,

		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				MarkdownDescription: machineResourceAttributes()["uuid"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id")),
				},
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: machineResourceAttributes()["id"].GetMarkdownDescription(),
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("uuid")),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: machineResourceAttributes()["name"].GetMarkdownDescription(),
				Computed:            true,
			},
			"machine_pool_id": schema.Int64Attribute{
				MarkdownDescription: machineResourceAttributes()["machine_pool_id"].GetMarkdownDescription(),
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the Machine",
				Computed:            true,
			},
			"labels": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: labelResourceAttributes()["key"].GetMarkdownDescription(),
						},
						"value": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: labelResourceAttributes()["value"].GetMarkdownDescription(),
						},
					},
				},
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

	var machine *client.Machine
	if data.ID.ValueInt64() != 0 {
		result, err := d.client.Machine().Get(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read machine by ID %d, got error: %s", data.ID.ValueInt64(), err))
			return
		}
		machine = result.Machine
	} else {
		result, err := d.client.Machine().List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read machines, got error: %s", err))
			return
		}

		for _, m := range result.Machines {
			if strings.EqualFold(data.UUID.ValueString(), m.UUID.String()) {
				machine = m
				break
			}
		}

		if machine == nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not find machine by UUID %s", data.UUID.ValueString()))
			return
		}
	}

	for _, label := range machine.Labels {
		data.Labels = append(data.Labels, LabelDataSourceModel{
			Key:   types.StringValue(label.Key),
			Value: types.StringValue(label.Value),
		})
	}

	data.ID = types.Int64Value(machine.ID)
	data.UUID = types.StringValue(machine.UUID.String())
	data.Name = types.StringValue(machine.Name)
	data.MachinePoolID = types.Int64Value(machine.MachinePoolID)
	data.Status = types.StringValue(machine.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

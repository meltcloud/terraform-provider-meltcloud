package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"terraform-provider-melt/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IPXEBootISOResource{}
var _ resource.ResourceWithImportState = &IPXEBootISOResource{}

func NewIPXEBootISOResource() resource.Resource {
	return &IPXEBootISOResource{}
}

// IPXEBootISOResource defines the resource implementation.
type IPXEBootISOResource struct {
	client *client.Client
}

// IPXEBootISOResourceModel describes the resource data model.
type IPXEBootISOResourceModel struct {
	ID          types.Int64       `tfsdk:"id"`
	ExpiresAt   timetypes.RFC3339 `tfsdk:"expires_at"`
	DownloadURL types.String      `tfsdk:"download_url"`
}

func (r *IPXEBootISOResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipxe_boot_iso"
}

func (r *IPXEBootISOResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "IPXEBootISO",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Melt ID of the ipxe boot ISO",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"expires_at": schema.StringAttribute{
				CustomType:          timetypes.RFC3339Type{},
				MarkdownDescription: "Timestamp when the ISO should expire",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"download_url": schema.StringAttribute{
				MarkdownDescription: "URL to download the ISO",
				Computed:            true,
			},
		},
	}
}

func (r *IPXEBootISOResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *IPXEBootISOResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IPXEBootISOResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	expiresAt, diagErr := data.ExpiresAt.ValueRFC3339Time()
	if diagErr != nil {
		resp.Diagnostics = diagErr
		return
	}

	ipxeBootISOCreateInput := &client.IPXEBootISOCreateInput{
		ExpiresAt: expiresAt.UTC(),
	}

	result, err := r.client.IPXEBootISO().Create(ctx, ipxeBootISOCreateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create ipxe boot iso, got error: %s", err))
		return
	}
	if result.Operation == nil {
		resp.Diagnostics.AddError("Server Error", "Created ipxe boot iso, but did not get operation")
		return
	}

	_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error during creation of ipxe boot iso, got error: %s", err))
		return
	}

	// TODO handle failed state

	_, err = r.client.IPXEBootISO().Get(ctx, result.IPXEBootISO.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ipxe boot iso, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.IPXEBootISO.ID)
	data.DownloadURL = types.StringValue(result.IPXEBootISO.DownloadURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IPXEBootISOResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IPXEBootISOResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.IPXEBootISO().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read cluster, got error: %s", err))
		return
	}

	data.DownloadURL = types.StringValue(result.IPXEBootISO.DownloadURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IPXEBootISOResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Resource Update Not Implemented", "ipxe_boot_iso does not support updates")
}

func (r *IPXEBootISOResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IPXEBootISOResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.IPXEBootISO().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete cluster, got error: %s", err))
		return
	}
}

func (r *IPXEBootISOResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

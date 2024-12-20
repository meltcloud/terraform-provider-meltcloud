package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"regexp"
	"strconv"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IPXEChainURLResource{}
var _ resource.ResourceWithImportState = &IPXEChainURLResource{}

func NewIPXEChainURLResource() resource.Resource {
	return &IPXEChainURLResource{}
}

// IPXEChainURLResource defines the resource implementation.
type IPXEChainURLResource struct {
	client *client.Client
}

// IPXEChainURLResourceModel describes the resource data model.
type IPXEChainURLResourceModel struct {
	ID        types.Int64       `tfsdk:"id"`
	Name      types.String      `tfsdk:"name"`
	ExpiresAt timetypes.RFC3339 `tfsdk:"expires_at"`
	URL       types.String      `tfsdk:"url"`
	Script    types.String      `tfsdk:"script"`
}

func (r *IPXEChainURLResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipxe_chain_url"
}

const iPXEChainURLDesc string = "Generate [iPXE Chain URLs](https://meltcloud.io/docs/guides/boot-config/create-ipxe-chain-urls.html) for providers that allow booting an iPXE Script or a remote iPXE URL (for example Equinix Metal)"

func iPXEChainURLResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the iPXE Chain URL on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the iPXE Chain URL, not case-sensitive. Must be unique within the organization.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"expires_at": schema.StringAttribute{
			CustomType:          timetypes.RFC3339Type{},
			MarkdownDescription: "Timestamp when the URL should expire",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"url": schema.StringAttribute{
			MarkdownDescription: "URL to the iPXE chain script",
			Computed:            true,
			Sensitive:           true,
		},
		"script": schema.StringAttribute{
			MarkdownDescription: "The complete iPXE script",
			Computed:            true,
			Sensitive:           true,
		},
	}
}

func (r *IPXEChainURLResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: iPXEChainURLDesc,

		Attributes: iPXEChainURLResourceAttributes(),
	}
}

func (r *IPXEChainURLResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IPXEChainURLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IPXEChainURLResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	expiresAt, diagErr := data.ExpiresAt.ValueRFC3339Time()
	if diagErr != nil {
		resp.Diagnostics = diagErr
		return
	}

	ipxeChainURLCreateInput := &client.IPXEChainURLCreateInput{
		Name:      data.Name.ValueString(),
		ExpiresAt: expiresAt.UTC(),
	}

	result, err := r.client.IPXEChainURL().Create(ctx, ipxeChainURLCreateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create ipxe chain url, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.IPXEChainURL.ID)
	data.URL = types.StringValue(result.IPXEChainURL.URL)
	data.Script = types.StringValue(result.IPXEChainURL.Script)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IPXEChainURLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IPXEChainURLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.IPXEChainURL().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ipxe chain url, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.IPXEChainURL.Name)
	data.ExpiresAt = timetypes.NewRFC3339TimeValue(result.IPXEChainURL.ExpiresAt.UTC())
	data.URL = types.StringValue(result.IPXEChainURL.URL)
	data.Script = types.StringValue(result.IPXEChainURL.Script)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IPXEChainURLResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Resource Update Not Implemented", "ipxe_chain_url does not support updates")
}

func (r *IPXEChainURLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IPXEChainURLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.IPXEChainURL().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete url, got error: %s", err))
		return
	}
}

var iPXEChainURLImportIDPattern = regexp.MustCompile(`ipxe_chain_urls/(\d+)`)

func (r *IPXEChainURLResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := iPXEChainURLImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", iPXEChainURLImportIDPattern.String()))
		return
	}

	id, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

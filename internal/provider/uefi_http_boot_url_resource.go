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
var _ resource.Resource = &UEFIHTTPBootURLResource{}
var _ resource.ResourceWithImportState = &UEFIHTTPBootURLResource{}

func NewUEFIHTTPBootURLResource() resource.Resource {
	return &UEFIHTTPBootURLResource{}
}

// UEFIHTTPBootURLResource defines the resource implementation.
type UEFIHTTPBootURLResource struct {
	client *client.Client
}

// UEFIHTTPBootURLResourceModel describes the resource data model.
type UEFIHTTPBootURLResourceModel struct {
	ID                 types.Int64 `tfsdk:"id"`
	IPXEBootArtifactID types.Int64 `tfsdk:"ipxe_boot_artifact_id"`

	Name          types.String      `tfsdk:"name"`
	ExpiresAt     timetypes.RFC3339 `tfsdk:"expires_at"`
	Protocols     types.String      `tfsdk:"protocols"`
	HTTPURLAMD64  types.String      `tfsdk:"http_url_amd64"`
	HTTPURLARM64  types.String      `tfsdk:"http_url_arm64"`
	HTTPSURLAMD64 types.String      `tfsdk:"https_url_amd64"`
	HTTPSURLARM64 types.String      `tfsdk:"https_url_arm64"`
}

func (r *UEFIHTTPBootURLResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_uefi_http_boot_url"
}

const uefiHTTPBootURLDesc string = "Generate [UEFI HTTP Boot URLs](https://meltcloud.io/docs/guides/boot-config/create-uefi-http-boot-urls.html) for servers that support UEFI HTTP Boot."

func uefiHTTPBootURLResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the UEFI HTTP Boot URL on meltcloud",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"ipxe_boot_artifact_id": schema.Int64Attribute{
			Required:            true,
			MarkdownDescription: "Internal ID of the iPXE Boot Artifact that this UEFI HTTP Boot URL should be generated for",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the UEFI HTTP Boot URL, not case-sensitive. Must be unique per iPXE Boot Artifact.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"protocols": schema.StringAttribute{
			MarkdownDescription: "Protocols to support. Must be either http_only, https_only or http_and_https.",
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
		"http_url_amd64": schema.StringAttribute{
			MarkdownDescription: "HTTP URL of the UEFI HTTP Boot URL for the amd64 architecture. Is null if protocols is set to https_only.",
			Computed:            true,
			Sensitive:           true,
		},
		"http_url_arm64": schema.StringAttribute{
			MarkdownDescription: "HTTP URL of the UEFI HTTP Boot URL for the arm64 architecture. Is null if protocols is set to https_only.",
			Computed:            true,
			Sensitive:           true,
		},
		"https_url_amd64": schema.StringAttribute{
			MarkdownDescription: "HTTPS URL of the UEFI HTTP Boot URL for the amd64 architecture. Is null if protocols is set to http_only.",
			Computed:            true,
			Sensitive:           true,
		},
		"https_url_arm64": schema.StringAttribute{
			MarkdownDescription: "HTTPS URL of the UEFI HTTP Boot URL for the arm64 architecture. Is null if protocols is set to http_only.",
			Computed:            true,
			Sensitive:           true,
		},
	}
}

func (r *UEFIHTTPBootURLResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: uefiHTTPBootURLDesc,

		Attributes: uefiHTTPBootURLResourceAttributes(),
	}
}

func (r *UEFIHTTPBootURLResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UEFIHTTPBootURLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UEFIHTTPBootURLResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	expiresAt, diagErr := data.ExpiresAt.ValueRFC3339Time()
	if diagErr != nil {
		resp.Diagnostics = diagErr
		return
	}

	createInput := &client.UEFIHTTPBootURLCreateInput{
		Name:      data.Name.ValueString(),
		Protocols: data.Protocols.ValueString(),
		ExpiresAt: expiresAt.UTC(),
	}

	result, err := r.client.UEFIHTTPBootURL().Create(ctx, data.IPXEBootArtifactID.ValueInt64(), createInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create boot url, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.UEFIHTTPBootURL.ID)
	data.HTTPURLAMD64 = types.StringValue(result.UEFIHTTPBootURL.HTTPURLAMD64)
	data.HTTPURLARM64 = types.StringValue(result.UEFIHTTPBootURL.HTTPURLARM64)
	data.HTTPSURLAMD64 = types.StringValue(result.UEFIHTTPBootURL.HTTPSURLAMD64)
	data.HTTPSURLARM64 = types.StringValue(result.UEFIHTTPBootURL.HTTPSURLARM64)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UEFIHTTPBootURLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UEFIHTTPBootURLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UEFIHTTPBootURL().Get(ctx, data.IPXEBootArtifactID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read boot url, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.UEFIHTTPBootURL.Name)
	data.Protocols = types.StringValue(result.UEFIHTTPBootURL.Protocols)
	data.ExpiresAt = timetypes.NewRFC3339TimeValue(result.UEFIHTTPBootURL.ExpiresAt.UTC())
	data.HTTPURLAMD64 = types.StringValue(result.UEFIHTTPBootURL.HTTPURLAMD64)
	data.HTTPURLARM64 = types.StringValue(result.UEFIHTTPBootURL.HTTPURLARM64)
	data.HTTPSURLAMD64 = types.StringValue(result.UEFIHTTPBootURL.HTTPSURLAMD64)
	data.HTTPSURLARM64 = types.StringValue(result.UEFIHTTPBootURL.HTTPSURLARM64)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UEFIHTTPBootURLResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Resource Update Not Implemented", "uefi_http_boot_url does not support updates")
}

func (r *UEFIHTTPBootURLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UEFIHTTPBootURLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UEFIHTTPBootURL().Delete(ctx, data.IPXEBootArtifactID.ValueInt64(), data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete url, got error: %s", err))
		return
	}
}

var uefiHTTPBootURLImportIDPattern = regexp.MustCompile(`ipxe_boot_artifacts/(\d+)/uefi_http_boot_urls/(\d+)`)

func (r *UEFIHTTPBootURLResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := uefiHTTPBootURLImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 3 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", uefiHTTPBootURLImportIDPattern.String()))
		return
	}

	ipxeBootArtifactID, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ipxe boot artifact ID: %s", err))
		return
	}

	id, err := strconv.ParseInt(match[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Invalid ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ipxe_boot_artifact_id"), ipxeBootArtifactID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

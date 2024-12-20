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
var _ resource.Resource = &IPXEBootArtifactResource{}
var _ resource.ResourceWithImportState = &IPXEBootArtifactResource{}

func NewIPXEBootArtifactResource() resource.Resource {
	return &IPXEBootArtifactResource{}
}

// IPXEBootArtifactResource defines the resource implementation.
type IPXEBootArtifactResource struct {
	client *client.Client
}

// IPXEBootArtifactResourceModel describes the resource data model.
type IPXEBootArtifactResourceModel struct {
	ID                  types.Int64       `tfsdk:"id"`
	Name                types.String      `tfsdk:"name"`
	ExpiresAt           timetypes.RFC3339 `tfsdk:"expires_at"`
	DownloadURLISO      types.String      `tfsdk:"download_url_iso"`
	DownloadURLPXE      types.String      `tfsdk:"download_url_pxe"`
	DownloadURLEFIAMD64 types.String      `tfsdk:"download_url_efi_amd64"`
	DownloadURLEFIARM64 types.String      `tfsdk:"download_url_efi_arm64"`
	DownloadURLRawAMD64 types.String      `tfsdk:"download_url_raw_amd64"`
}

func (r *IPXEBootArtifactResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipxe_boot_artifact"
}

const iPXEBootArtifactDesc string = "An [iPXE Boot Artifact](https://meltcloud.io/docs/guides/boot-config/create-ipxe-boot-artifacts.html) contains a set of bootable images with an X509 client certificate to securely boot into your meltcloud organization."

func iPXEBootArtifactResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the iPXE Boot Artifact",
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: "Name of the iPXE Boot Artifact, not case-sensitive. Must be unique within the organization.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"expires_at": schema.StringAttribute{
			CustomType:          timetypes.RFC3339Type{},
			MarkdownDescription: "Timestamp when the artifact should expire",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"download_url_iso": schema.StringAttribute{
			MarkdownDescription: "URL to download the ISO",
			Computed:            true,
			Sensitive:           true,
		},
		"download_url_pxe": schema.StringAttribute{
			MarkdownDescription: "URL to download the PCBIOS artifact (.undionly)",
			Computed:            true,
			Sensitive:           true,
		},
		"download_url_efi_amd64": schema.StringAttribute{
			MarkdownDescription: "URL to download the amd64 EFI boot artifact",
			Computed:            true,
			Sensitive:           true,
		},
		"download_url_efi_arm64": schema.StringAttribute{
			MarkdownDescription: "URL to download the arm64 EFI boot artifact",
			Computed:            true,
			Sensitive:           true,
		},
		"download_url_raw_amd64": schema.StringAttribute{
			MarkdownDescription: "URL to download the amd64 Raw boot artifact",
			Computed:            true,
			Sensitive:           true,
		},
	}
}

func (r *IPXEBootArtifactResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: iPXEBootArtifactDesc,

		Attributes: iPXEBootArtifactResourceAttributes(),
	}
}

func (r *IPXEBootArtifactResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IPXEBootArtifactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IPXEBootArtifactResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	name := data.Name.ValueString()
	expiresAt, diagErr := data.ExpiresAt.ValueRFC3339Time()
	if diagErr != nil {
		resp.Diagnostics = diagErr
		return
	}

	iPXEBootArtifactCreateInput := &client.IPXEBootArtifactCreateInput{
		ExpiresAt: expiresAt.UTC(),
		Name:      name,
	}

	result, err := r.client.IPXEBootArtifact().Create(ctx, iPXEBootArtifactCreateInput)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create ipxe boot artifact, got error: %s", err))
		return
	}
	if result.Operation == nil {
		resp.Diagnostics.AddError("Server Error", "Created ipxe boot artifact, but did not get operation")
		return
	}

	_, err = r.client.Operation().PollUntilDone(ctx, result.Operation.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("error during creation of ipxe boot artifact, got error: %s", err))
		return
	}

	result, err = r.client.IPXEBootArtifact().Get(ctx, result.IPXEBootArtifact.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ipxe boot artifact, got error: %s", err))
		return
	}

	data.ID = types.Int64Value(result.IPXEBootArtifact.ID)
	data.DownloadURLISO = types.StringValue(result.IPXEBootArtifact.DownloadURLISO)
	data.DownloadURLPXE = types.StringValue(result.IPXEBootArtifact.DownloadURLPXE)
	data.DownloadURLEFIAMD64 = types.StringValue(result.IPXEBootArtifact.DownloadURLEFIAMD64)
	data.DownloadURLEFIARM64 = types.StringValue(result.IPXEBootArtifact.DownloadURLEFIARM64)
	data.DownloadURLRawAMD64 = types.StringValue(result.IPXEBootArtifact.DownloadURLRawAMD64)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IPXEBootArtifactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IPXEBootArtifactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.IPXEBootArtifact().Get(ctx, data.ID.ValueInt64())
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read ipxe boot artifact, got error: %s", err))
		return
	}

	data.Name = types.StringValue(result.IPXEBootArtifact.Name)
	data.ExpiresAt = timetypes.NewRFC3339TimeValue(result.IPXEBootArtifact.ExpiresAt.UTC())
	data.DownloadURLISO = types.StringValue(result.IPXEBootArtifact.DownloadURLISO)
	data.DownloadURLPXE = types.StringValue(result.IPXEBootArtifact.DownloadURLPXE)
	data.DownloadURLEFIAMD64 = types.StringValue(result.IPXEBootArtifact.DownloadURLEFIAMD64)
	data.DownloadURLEFIARM64 = types.StringValue(result.IPXEBootArtifact.DownloadURLEFIARM64)
	data.DownloadURLRawAMD64 = types.StringValue(result.IPXEBootArtifact.DownloadURLRawAMD64)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IPXEBootArtifactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Resource Update Not Implemented", "ipxe_boot_artifact does not support updates")
}

func (r *IPXEBootArtifactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IPXEBootArtifactResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.IPXEBootArtifact().Delete(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete cluster, got error: %s", err))
		return
	}
}

var iPXEBootArtifactImportIDPattern = regexp.MustCompile(`ipxe_boot_artifacts/(\d+)`)

func (r *IPXEBootArtifactResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	match := iPXEBootArtifactImportIDPattern.FindStringSubmatch(req.ID)
	if len(match) != 2 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ID does not follow format: %s", iPXEBootArtifactImportIDPattern.String()))
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

package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &NestResource{}
	_ resource.ResourceWithImportState = &NestResource{}
)

func NewNestResource() resource.Resource {
	return &NestResource{}
}

type NestResource struct {
	client *client.Client
}

type NestResourceModel struct {
	ID      types.Int64  `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Status  types.String `tfsdk:"status"`
	Version types.String `tfsdk:"version"`
}

func (r *NestResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nest"
}

const nestDesc = "The [Nest](https://docs.meltcloud.io/concepts/nest/overview) is the management appliance that hosts the Foundry and the Kubernetes control planes. " +
	"Since it's created automatically during the Nest installation, it can't be created by Terraform and must be [imported](https://developer.hashicorp.com/terraform/language/import). " +
	"Same for deletion: to stop managing it, use a [`removed`](https://developer.hashicorp.com/terraform/language/resources/syntax#removing-resources) block with `destroy = false`.\n\n" +
	"Needs the API Key of the admin Organization.\n\n" +
	"~> Changing `version` updates and reboots one Nest Instance after the other. The Foundry's API can be briefly unavailable during the update."

func (r *NestResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: nestDesc,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Internal ID of the Nest on meltcloud",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the Nest, given at install",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the Nest, for example `ready` or `updating`",
			},
			"version": optionalComputedString("The Nest version to run, for example `1.35.3-nest.67`. It has to be a version this Nest can update to. Leave it out to keep the running one"),
		},
	}
}

func (r *NestResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// The Nest exists from the install on, so it is imported, never created.
func (r *NestResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Nest Must Be Imported",
		"The Nest installer creates the Nest, so Terraform cannot create one. Import it instead:\n\n"+
			"import {\n  to = meltcloud_nest.<name>\n  id = \"nest\"\n}")
}

func (r *NestResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NestResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Nest().Get(ctx)
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Nest, got error: %s", err))
		return
	}

	setNestValues(result.Nest, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NestResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NestResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.apply(ctx, &data, resp.Diagnostics.AddError)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NestResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Nest Cannot Be Deleted",
		"To stop managing the Nest with Terraform, replace the resource with:\n\n"+
			"removed {\n  from = meltcloud_nest.<name>\n  lifecycle {\n    destroy = false\n  }\n}")
}

func (r *NestResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "nest" {
		resp.Diagnostics.AddError("Client Error", "ID does not follow format: nest")
		return
	}

	result, err := r.client.Nest().Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Nest, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), result.Nest.ID)...)
}

// Rolls the Nest out to the planned version, if it has one, and reads the Nest back into data.
func (r *NestResource) apply(ctx context.Context, data *NestResourceModel, addError func(string, string)) {
	if version := knownString(data.Version); version != nil {
		result, err := r.client.Nest().Update(ctx, *version)
		if err != nil {
			addError("Client Error", fmt.Sprintf("Unable to update Nest, got error: %s", err))
			return
		}

		if result.Operation != nil {
			if _, err := r.client.Operation().PollUntilDoneThroughRestarts(ctx, result.Operation.ID); err != nil {
				addError("Client Error", fmt.Sprintf("Unable to update Nest, got error: %s", err))
				return
			}
		}
	}

	result, err := r.client.Nest().Get(ctx)
	if err != nil {
		addError("Client Error", fmt.Sprintf("Unable to read Nest, got error: %s", err))
		return
	}
	setNestValues(result.Nest, data)
}

func setNestValues(nest *client.Nest, data *NestResourceModel) {
	data.ID = types.Int64Value(nest.ID)
	data.Name = types.StringValue(nest.Name)
	data.Status = types.StringValue(nest.Status)
	data.Version = types.StringValue(nest.Version)
}

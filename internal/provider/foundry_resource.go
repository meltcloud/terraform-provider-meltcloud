package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &FoundryResource{}
	_ resource.ResourceWithImportState = &FoundryResource{}
)

func NewFoundryResource() resource.Resource {
	return &FoundryResource{}
}

type FoundryResource struct {
	client *client.Client
}

type FoundryResourceModel struct {
	ID                 types.Int64  `tfsdk:"id"`
	Status             types.String `tfsdk:"status"`
	Version            types.String `tfsdk:"version"`
	HATopology         types.String `tfsdk:"ha_topology"`
	RailsLogLevel      types.String `tfsdk:"rails_log_level"`
	IdentityProvider   types.String `tfsdk:"identity_provider"`
	Auth0Domain        types.String `tfsdk:"auth0_domain"`
	Auth0ClientID      types.String `tfsdk:"auth0_client_id"`
	Auth0ClientSecret  types.String `tfsdk:"auth0_client_secret"`
	SentryEnabled      types.Bool   `tfsdk:"sentry_enabled"`
	SentryDSN          types.String `tfsdk:"sentry_dsn"`
	SentryEnvironment  types.String `tfsdk:"sentry_environment"`
	LetsencryptEnabled types.Bool   `tfsdk:"letsencrypt_enabled"`
	LetsencryptEmail   types.String `tfsdk:"letsencrypt_email"`
	LetsencryptStaging types.Bool   `tfsdk:"letsencrypt_staging"`
}

func (r *FoundryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_foundry"
}

const foundryDesc = "The [Foundry](https://docs.meltcloud.io/concepts/nest/overview#foundry) is the meltcloud management plane that runs on the Nest. " +
	"The Nest installer creates it, so it is [imported](https://developer.hashicorp.com/terraform/language/import) rather than created, and it cannot be destroyed: " +
	"to stop managing it, use a [`removed`](https://developer.hashicorp.com/terraform/language/resources/syntax#removing-resources) block with `destroy = false`.\n\n" +
	"Needs the API Key of the admin Organization.\n\n" +
	"~> Changing `version` updates the Foundry. Its API can be briefly unavailable during the update, and the apply waits until the update has finished.\n\n" +
	"~> An apply also rolls out changes saved in the admin area but not rolled out yet. The plan does not show them."

func optionalComputedString(description string, validators ...validator.String) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: description,
		Optional:            true,
		Computed:            true,
		Validators:          validators,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func optionalComputedBool(description string) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: description,
		Optional:            true,
		Computed:            true,
		PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
	}
}

func secretString(description string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: description + ". The Foundry never returns it, so Terraform cannot see a change made outside of it",
		Optional:            true,
		Sensitive:           true,
	}
}

func foundryResourceAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed:            true,
			MarkdownDescription: "Internal ID of the Foundry on meltcloud",
			PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
		},
		"status": schema.StringAttribute{
			Computed:            true,
			MarkdownDescription: "Status of the Foundry, for example `ready` or `updating`",
		},
		"version": optionalComputedString("The Foundry version to run, for example `1.5.0`. It has to be a version this Foundry can update to. Leave it out to keep the running one"),
		"ha_topology": schema.StringAttribute{
			Computed: true,
			MarkdownDescription: "[High availability](https://docs.meltcloud.io/concepts/nest/high-availability) mode of the Foundry: `none` for Single Instance, `host` for Multi-instance, `zone` for Multi-zone. " +
				"[Enable high availability](https://docs.meltcloud.io/tasks/foundry/enable-high-availability) in the admin area",
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"rails_log_level":   optionalComputedString("Log level of the Foundry application (web and worker)", stringvalidator.OneOf("debug", "info", "warn", "error", "fatal", "unknown")),
		"identity_provider": optionalComputedString("How users sign in: `auth0` through an Auth0 tenant, or `dummy`, which signs everyone in as admin, for trials only", stringvalidator.OneOf("auth0", "dummy")),
		"auth0_domain":      optionalComputedString("Domain of the Auth0 tenant, for example `tenant.eu.auth0.com`"),
		"auth0_client_id":   optionalComputedString("Client ID of the Auth0 application"),
		"auth0_client_secret": secretString(
			"Client secret of the Auth0 application",
		),
		"sentry_enabled":     optionalComputedBool("Report errors to Sentry"),
		"sentry_dsn":         secretString("DSN errors are reported to"),
		"sentry_environment": optionalComputedString("Environment the errors are reported under"),
		"letsencrypt_enabled": optionalComputedBool(
			"Issue the certificates for the Foundry hostnames with Let's Encrypt's HTTP01 challenge. Port 80 of the Foundry VIP must be reachable from the internet",
		),
		"letsencrypt_email":   optionalComputedString("Email address Let's Encrypt sends expiry notices to"),
		"letsencrypt_staging": optionalComputedBool("Use the Let's Encrypt staging environment, whose certificates browsers do not trust"),
	}
}

func (r *FoundryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: foundryDesc + "\n\n" +
			"A setting left out keeps its value, and an empty string clears it.",
		Attributes: foundryResourceAttributes(),
	}
}

func (r *FoundryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// The Foundry exists from the install on, so it is imported, never created.
func (r *FoundryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Foundry Must Be Imported",
		"The Nest installer creates the Foundry, so Terraform cannot create one. Import it instead:\n\n"+
			"import {\n  to = meltcloud_foundry.<name>\n  id = \"foundry\"\n}")
}

func (r *FoundryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FoundryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Foundry().Get(ctx)
	if err != nil {
		if err.HTTPStatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Foundry, got error: %s", err))
		return
	}

	setFoundryValues(result.Foundry, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FoundryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FoundryResourceModel
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

func (r *FoundryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError("Foundry Cannot Be Deleted",
		"The Foundry runs the Nest. To stop managing it with Terraform, replace the resource with:\n\n"+
			"removed {\n  from = meltcloud_foundry.<name>\n  lifecycle {\n    destroy = false\n  }\n}")
}

func (r *FoundryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != "foundry" {
		resp.Diagnostics.AddError("Client Error", "ID does not follow format: foundry")
		return
	}

	result, err := r.client.Foundry().Get(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Foundry, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), result.Foundry.ID)...)
}

// Sends the settings and version, waits for the update, and reads the Foundry back into data.
func (r *FoundryResource) apply(ctx context.Context, data *FoundryResourceModel, addError func(string, string)) {
	result, err := r.client.Foundry().Update(ctx, foundryUpdateInput(data))
	if err != nil {
		addError("Client Error", fmt.Sprintf("Unable to update Foundry, got error: %s", err))
		return
	}
	if !r.await(ctx, result.Operation, addError) {
		return
	}

	result, err = r.client.Foundry().Get(ctx)
	if err != nil {
		addError("Client Error", fmt.Sprintf("Unable to read Foundry, got error: %s", err))
		return
	}
	setFoundryValues(result.Foundry, data)
}

func (r *FoundryResource) await(ctx context.Context, operation *client.Operation, addError func(string, string)) bool {
	if operation == nil {
		return true
	}

	if _, err := r.client.Operation().PollUntilDoneThroughRestarts(ctx, operation.ID); err != nil {
		addError("Client Error", fmt.Sprintf("Unable to update Foundry, got error: %s", err))
		return false
	}
	return true
}

func knownString(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return value.ValueStringPointer()
}

func knownBool(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return value.ValueBoolPointer()
}

func foundryUpdateInput(data *FoundryResourceModel) *client.FoundryUpdateInput {
	return &client.FoundryUpdateInput{
		Version:            knownString(data.Version),
		RailsLogLevel:      knownString(data.RailsLogLevel),
		IdentityProvider:   knownString(data.IdentityProvider),
		Auth0Domain:        knownString(data.Auth0Domain),
		Auth0ClientID:      knownString(data.Auth0ClientID),
		Auth0ClientSecret:  knownString(data.Auth0ClientSecret),
		SentryEnabled:      knownBool(data.SentryEnabled),
		SentryDSN:          knownString(data.SentryDSN),
		SentryEnvironment:  knownString(data.SentryEnvironment),
		LetsencryptEnabled: knownBool(data.LetsencryptEnabled),
		LetsencryptEmail:   knownString(data.LetsencryptEmail),
		LetsencryptStaging: knownBool(data.LetsencryptStaging),
	}
}

// The secrets stay as they are in data: the Foundry does not return them.
func setFoundryValues(foundry *client.Foundry, data *FoundryResourceModel) {
	data.ID = types.Int64Value(foundry.ID)
	data.Status = types.StringValue(foundry.Status)
	data.Version = types.StringValue(foundry.Version)
	data.HATopology = types.StringValue(foundry.HATopology)
	data.RailsLogLevel = types.StringValue(foundry.RailsLogLevel)
	data.IdentityProvider = types.StringValue(foundry.IdentityProvider)
	data.Auth0Domain = types.StringValue(foundry.Auth0Domain)
	data.Auth0ClientID = types.StringValue(foundry.Auth0ClientID)
	data.SentryEnabled = types.BoolValue(foundry.SentryEnabled)
	data.SentryEnvironment = types.StringValue(foundry.SentryEnvironment)
	data.LetsencryptEnabled = types.BoolValue(foundry.LetsencryptEnabled)
	data.LetsencryptEmail = types.StringValue(foundry.LetsencryptEmail)
	data.LetsencryptStaging = types.BoolValue(foundry.LetsencryptStaging)
}

package provider

import (
	"context"
	"fmt"
	"terraform-provider-meltcloud/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	ID                types.Int64              `tfsdk:"id"`
	Status            types.String             `tfsdk:"status"`
	Version           types.String             `tfsdk:"version"`
	HATopology        types.String             `tfsdk:"ha_topology"`
	RailsLogLevel     types.String             `tfsdk:"rails_log_level"`
	IdentityProvider  types.String             `tfsdk:"identity_provider"`
	Auth0Config       *FoundryAuth0Model       `tfsdk:"auth0_config"`
	LetsencryptConfig *FoundryLetsencryptModel `tfsdk:"letsencrypt_config"`
	SentryConfig      *FoundrySentryModel      `tfsdk:"sentry_config"`
}

type FoundryAuth0Model struct {
	Domain                types.String `tfsdk:"domain"`
	ClientID              types.String `tfsdk:"client_id"`
	ClientSecretWO        types.String `tfsdk:"client_secret_wo"`
	ClientSecretWOVersion types.Int64  `tfsdk:"client_secret_wo_version"`
}

type FoundryLetsencryptModel struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	Email   types.String `tfsdk:"email"`
	Staging types.Bool   `tfsdk:"staging"`
}

type FoundrySentryModel struct {
	Enabled      types.Bool   `tfsdk:"enabled"`
	DSNWO        types.String `tfsdk:"dsn_wo"`
	DSNWOVersion types.Int64  `tfsdk:"dsn_wo_version"`
	Environment  types.String `tfsdk:"environment"`
}

func (r *FoundryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_foundry"
}

const foundryDesc = "The [Foundry](https://docs.meltcloud.io/concepts/nest/overview#foundry) is the meltcloud management plane that runs on the Nest. " +
	"Since it's created automatically during the Nest installation, it can't be created by Terraform and must be [imported](https://developer.hashicorp.com/terraform/language/import). " +
	"Same for deletion: to stop managing it, use a [`removed`](https://developer.hashicorp.com/terraform/language/resources/syntax#removing-resources) block with `destroy = false`.\n\n" +
	"Needs the API Key of the admin Organization.\n\n" +
	"~> Changing `version` updates the Foundry. Its API can be briefly unavailable during the update."

func optionalComputedString(description string, validators ...validator.String) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: description,
		Optional:            true,
		Computed:            true,
		Validators:          validators,
		PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
}

func settingString(description string) schema.StringAttribute {
	return schema.StringAttribute{MarkdownDescription: description, Optional: true}
}

func writeOnlySecret(attributes map[string]schema.Attribute, name, description, onChange string) {
	version := name + "_version"
	attributes[name] = schema.StringAttribute{
		MarkdownDescription: description + ". Write-only, needs Terraform 1.11 or later: it is not stored in the state. " + onChange,
		Optional:            true,
		Sensitive:           true,
		WriteOnly:           true,
		Validators:          []validator.String{stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName(version))},
	}
	attributes[version] = schema.Int64Attribute{
		MarkdownDescription: "Version of `" + name + "`, see there",
		Optional:            true,
		Validators:          []validator.Int64{int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName(name))},
	}
}

func foundryAuth0Attribute() schema.SingleNestedAttribute {
	attributes := map[string]schema.Attribute{
		"domain":    schema.StringAttribute{Required: true, MarkdownDescription: "Domain of the Auth0 tenant, for example `tenant.eu.auth0.com`"},
		"client_id": schema.StringAttribute{Required: true, MarkdownDescription: "Client ID of the Auth0 application"},
	}
	writeOnlySecret(attributes, "client_secret_wo", "Client secret of the Auth0 application", "Change `client_secret_wo_version` to send a new one")
	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "The Auth0 tenant users sign in through with `identity_provider = \"auth0\"`. Leave it out to keep the current settings",
		Attributes:          attributes,
	}
}

func foundryLetsencryptAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "Use certificates from Let's Encrypt. Leave it out to keep the current settings",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Required:            true,
				MarkdownDescription: "Issue the certificates for the Foundry hostnames with Let's Encrypt's HTTP01 challenge. Port 80 of the Foundry VIP must be reachable from the internet",
			},
			"email": settingString("Email address Let's Encrypt sends expiry notices to"),
			"staging": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Use the Let's Encrypt staging environment, whose certificates browsers do not trust. Defaults to `false`",
			},
		},
	}
}

func foundrySentryAttribute() schema.SingleNestedAttribute {
	attributes := map[string]schema.Attribute{
		"enabled":     schema.BoolAttribute{Required: true, MarkdownDescription: "Report errors of the Foundry application to Sentry"},
		"environment": settingString("Environment the errors are reported under, for example `production`"),
	}
	writeOnlySecret(attributes, "dsn_wo", "DSN of the Sentry project", "Change `dsn_wo_version` to send a new one")
	return schema.SingleNestedAttribute{
		Optional:            true,
		MarkdownDescription: "Error reporting to Sentry. Leave it out to keep the current settings",
		Attributes:          attributes,
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
		"rails_log_level":    optionalComputedString("Log level of the Foundry application (web and worker)", stringvalidator.OneOf("debug", "info", "warn", "error", "fatal", "unknown")),
		"identity_provider":  optionalComputedString("How users sign in: `auth0` through an Auth0 tenant, or `dummy`, which signs everyone in as admin, for trials only", stringvalidator.OneOf("auth0", "dummy")),
		"auth0_config":       foundryAuth0Attribute(),
		"letsencrypt_config": foundryLetsencryptAttribute(),
		"sentry_config":      foundrySentryAttribute(),
	}
}

func (r *FoundryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: foundryDesc,
		Attributes:          foundryResourceAttributes(),
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
	var data, config, state FoundryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.apply(ctx, &data, foundryUpdateInput(&data, &config, &state), resp.Diagnostics.AddError)
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

	// The blocks of the settings in use are filled in, so a config that matches the Foundry plans no change.
	foundry := result.Foundry
	var data FoundryResourceModel
	if foundry.Auth0Domain != "" {
		data.Auth0Config = &FoundryAuth0Model{}
	}
	if foundry.LetsencryptEnabled {
		data.LetsencryptConfig = &FoundryLetsencryptModel{}
	}
	if foundry.SentryEnabled {
		data.SentryConfig = &FoundrySentryModel{}
	}
	setFoundryValues(foundry, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Sends the settings and version, waits for the update, and reads the Foundry back into data.
func (r *FoundryResource) apply(ctx context.Context, data *FoundryResourceModel, input *client.FoundryUpdateInput, addError func(string, string)) {
	result, err := r.client.Foundry().Update(ctx, input)
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

	if _, err := r.client.Operation().PollUntilDone(ctx, operation.ID); err != nil {
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

// A block left out sends nothing. A secret goes along only when its version changed.
func foundryUpdateInput(data, config, state *FoundryResourceModel) *client.FoundryUpdateInput {
	input := &client.FoundryUpdateInput{
		Version:          knownString(data.Version),
		RailsLogLevel:    knownString(data.RailsLogLevel),
		IdentityProvider: knownString(data.IdentityProvider),
	}
	if auth0 := data.Auth0Config; auth0 != nil {
		input.Auth0Domain = auth0.Domain.ValueStringPointer()
		input.Auth0ClientID = auth0.ClientID.ValueStringPointer()
		if config.Auth0Config != nil && (state.Auth0Config == nil || !auth0.ClientSecretWOVersion.Equal(state.Auth0Config.ClientSecretWOVersion)) {
			input.Auth0ClientSecret = knownString(config.Auth0Config.ClientSecretWO)
		}
	}
	if letsencrypt := data.LetsencryptConfig; letsencrypt != nil {
		input.LetsencryptEnabled = letsencrypt.Enabled.ValueBoolPointer()
		input.LetsencryptEmail = clearedIfNull(letsencrypt.Email)
		input.LetsencryptStaging = knownBool(letsencrypt.Staging)
	}
	if sentry := data.SentryConfig; sentry != nil {
		input.SentryEnabled = sentry.Enabled.ValueBoolPointer()
		input.SentryEnvironment = clearedIfNull(sentry.Environment)
		if config.SentryConfig != nil && (state.SentryConfig == nil || !sentry.DSNWOVersion.Equal(state.SentryConfig.DSNWOVersion)) {
			input.SentryDSN = knownString(config.SentryConfig.DSNWO)
		}
	}
	return input
}

// Fills in only the blocks data has, so a block left out of the config stays unmanaged.
func setFoundryValues(foundry *client.Foundry, data *FoundryResourceModel) {
	data.ID = types.Int64Value(foundry.ID)
	data.Status = types.StringValue(foundry.Status)
	data.Version = types.StringValue(foundry.Version)
	data.HATopology = types.StringValue(foundry.HATopology)
	data.RailsLogLevel = types.StringValue(foundry.RailsLogLevel)
	data.IdentityProvider = types.StringValue(foundry.IdentityProvider)
	if auth0 := data.Auth0Config; auth0 != nil {
		auth0.Domain = types.StringValue(foundry.Auth0Domain)
		auth0.ClientID = types.StringValue(foundry.Auth0ClientID)
		auth0.ClientSecretWO = types.StringNull()
	}
	if letsencrypt := data.LetsencryptConfig; letsencrypt != nil {
		letsencrypt.Enabled = types.BoolValue(foundry.LetsencryptEnabled)
		letsencrypt.Email = emptyAsNull(foundry.LetsencryptEmail)
		letsencrypt.Staging = types.BoolValue(foundry.LetsencryptStaging)
	}
	if sentry := data.SentryConfig; sentry != nil {
		sentry.Enabled = types.BoolValue(foundry.SentryEnabled)
		sentry.Environment = emptyAsNull(foundry.SentryEnvironment)
		sentry.DSNWO = types.StringNull()
	}
}

// An empty string makes the Foundry clear the setting.
func clearedIfNull(value types.String) *string {
	cleared := value.ValueString()
	return &cleared
}

func emptyAsNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

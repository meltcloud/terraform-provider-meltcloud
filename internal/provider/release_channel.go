package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const releaseChannelManual = "manual"

var bundleMinorPattern = regexp.MustCompile(`^(\d+\.\d+)\.\d+`)

// bundleMinor returns the Kubernetes minor version of a Kubernetes Bundle name, e.g. 1.35 for 1.35.5-melt.30.
func bundleMinor(bundle string) (string, bool) {
	match := bundleMinorPattern.FindStringSubmatch(bundle)
	if match == nil {
		return "", false
	}
	return match[1], true
}

func versionAttribute(desc string) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc + ". Required unless `release_channel` is `manual`, in which case it is derived from `manual_version` if not set.",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			versionFromManualVersion(),
		},
	}
}

func releaseChannelAttribute(extraDesc string) schema.StringAttribute {
	desc := "Release channel the Kubernetes Bundle is taken from, e.g. `stable`. The available channels vary per installation. " +
		"Set to `manual` to pin the Kubernetes Bundle given in `manual_version`."
	if extraDesc != "" {
		desc += " " + extraDesc
	}
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Required:            true,
	}
}

func manualVersionAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Kubernetes Bundle to pin, e.g. `1.35.5-melt.30`. Required if `release_channel` is `manual`, must not be set otherwise.",
		Optional:            true,
	}
}

func kubernetesBundleAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "Name of the Kubernetes Bundle currently applied, e.g. `1.35.5-melt.30`",
		Computed:            true,
	}
}

// releaseChannelValidator runs validateReleaseChannel on a resource's release_channel, manual_version and
// version attributes.
type releaseChannelValidator struct{}

var _ resource.ConfigValidator = releaseChannelValidator{}

func (v releaseChannelValidator) Description(ctx context.Context) string {
	return "manual_version is set exactly when release_channel is manual, and version matches its minor."
}

func (v releaseChannelValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v releaseChannelValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var releaseChannel, manualVersion, version types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("release_channel"), &releaseChannel)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("manual_version"), &manualVersion)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("version"), &version)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(validateReleaseChannel(releaseChannel, manualVersion, version)...)
}

// validateReleaseChannel checks that manual_version is set exactly when release_channel is manual, and that
// version is either set or derivable from manual_version and matches its minor.
func validateReleaseChannel(releaseChannel, manualVersion, version types.String) diag.Diagnostics {
	var diags diag.Diagnostics

	if releaseChannel.IsUnknown() || manualVersion.IsUnknown() {
		return diags
	}

	if releaseChannel.ValueString() != releaseChannelManual {
		if !manualVersion.IsNull() {
			diags.AddAttributeError(
				path.Root("manual_version"),
				"Invalid Attribute Combination",
				"manual_version can only be set if release_channel is \"manual\".",
			)
		}
		if version.IsNull() {
			diags.AddAttributeError(
				path.Root("version"),
				"Missing Attribute Configuration",
				"version must be set unless release_channel is \"manual\".",
			)
		}
		return diags
	}

	if manualVersion.IsNull() {
		diags.AddAttributeError(
			path.Root("manual_version"),
			"Missing Attribute Configuration",
			"manual_version must be set if release_channel is \"manual\".",
		)
		return diags
	}

	minor, ok := bundleMinor(manualVersion.ValueString())
	if !ok {
		diags.AddAttributeError(
			path.Root("manual_version"),
			"Invalid Attribute Value",
			"manual_version must be a Kubernetes Bundle name, e.g. \"1.35.5-melt.30\".",
		)
		return diags
	}

	if !version.IsNull() && !version.IsUnknown() && version.ValueString() != minor {
		diags.AddAttributeError(
			path.Root("version"),
			"Invalid Attribute Combination",
			fmt.Sprintf("version must match the minor version of manual_version (%s), or be omitted.", minor),
		)
	}

	return diags
}

func versionFromManualVersion() planmodifier.String {
	return versionFromManualVersionModifier{}
}

// versionFromManualVersionModifier plans an unconfigured version as the minor of manual_version, so pinning a
// bundle shows the resulting version in the plan rather than a perpetual "known after apply".
type versionFromManualVersionModifier struct{}

func (m versionFromManualVersionModifier) Description(ctx context.Context) string {
	return "If not configured, the version is derived from manual_version."
}

func (m versionFromManualVersionModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m versionFromManualVersionModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.ConfigValue.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var manualVersion types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("manual_version"), &manualVersion)...)
	if resp.Diagnostics.HasError() || manualVersion.IsNull() || manualVersion.IsUnknown() {
		return
	}

	if minor, ok := bundleMinor(manualVersion.ValueString()); ok {
		resp.PlanValue = types.StringValue(minor)
	}
}

func optionalString(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	return value.ValueStringPointer()
}

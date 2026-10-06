package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBundleMinor(t *testing.T) {
	cases := map[string]string{
		"1.35.5-melt.30": "1.35",
		"1.9.0":          "1.9",
		"1.35":           "",
		"latest":         "",
	}
	for bundle, want := range cases {
		got, ok := bundleMinor(bundle)
		if ok != (want != "") || got != want {
			t.Errorf("bundleMinor(%q) = %q, %t; want %q", bundle, got, ok, want)
		}
	}
}

func TestValidateReleaseChannel(t *testing.T) {
	null := types.StringNull()
	unknown := types.StringUnknown()
	str := types.StringValue

	cases := []struct {
		name           string
		releaseChannel types.String
		manualVersion  types.String
		version        types.String
		errorPaths     []path.Path
	}{
		{"channel with version", str("stable"), null, str("1.35"), nil},
		{"default channel with version", null, null, str("1.35"), nil},
		{"channel without version", str("stable"), null, null, []path.Path{path.Root("version")}},
		{"channel with pin", str("stable"), str("1.35.5-melt.30"), str("1.35"), []path.Path{path.Root("manual_version")}},
		{"default channel with pin", null, str("1.35.5-melt.30"), str("1.35"), []path.Path{path.Root("manual_version")}},
		{"manual with pin", str("manual"), str("1.35.5-melt.30"), null, nil},
		{"manual with pin and matching version", str("manual"), str("1.35.5-melt.30"), str("1.35"), nil},
		{"manual with pin and other version", str("manual"), str("1.35.5-melt.30"), str("1.34"), []path.Path{path.Root("version")}},
		{"manual with pin and unknown version", str("manual"), str("1.35.5-melt.30"), unknown, nil},
		{"manual with invalid pin", str("manual"), str("1.35"), null, []path.Path{path.Root("manual_version")}},
		{"manual without pin", str("manual"), null, str("1.35"), []path.Path{path.Root("manual_version")}},
		{"unknown channel", unknown, null, null, nil},
		{"unknown pin", str("stable"), unknown, null, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			diags := validateReleaseChannel(c.releaseChannel, c.manualVersion, c.version)

			var errorPaths []path.Path
			for _, d := range diags.Errors() {
				if withPath, ok := d.(interface{ Path() path.Path }); ok {
					errorPaths = append(errorPaths, withPath.Path())
				}
			}

			if len(errorPaths) != len(c.errorPaths) {
				t.Fatalf("got errors at %v, want %v", errorPaths, c.errorPaths)
			}
			for i := range errorPaths {
				if !errorPaths[i].Equal(c.errorPaths[i]) {
					t.Errorf("got errors at %v, want %v", errorPaths, c.errorPaths)
				}
			}
		})
	}
}

func TestVersionFromManualVersion(t *testing.T) {
	ctx := context.Background()
	s := schema.Schema{Attributes: map[string]schema.Attribute{
		"version":        versionAttribute("Version"),
		"manual_version": manualVersionAttribute(),
	}}

	cases := []struct {
		name          string
		config        types.String
		manualVersion tftypes.Value
		want          types.String
	}{
		{"derived from pin", types.StringNull(), tftypes.NewValue(tftypes.String, "1.35.5-melt.30"), types.StringValue("1.35")},
		{"configured", types.StringValue("1.35"), tftypes.NewValue(tftypes.String, "1.35.5-melt.30"), types.StringUnknown()},
		{"unknown pin", types.StringNull(), tftypes.NewValue(tftypes.String, tftypes.UnknownValue), types.StringUnknown()},
		{"no pin", types.StringNull(), tftypes.NewValue(tftypes.String, nil), types.StringUnknown()},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plan := tfsdk.Plan{
				Schema: s,
				Raw: tftypes.NewValue(s.Type().TerraformType(ctx), map[string]tftypes.Value{
					"version":        tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
					"manual_version": c.manualVersion,
				}),
			}
			req := planmodifier.StringRequest{Path: path.Root("version"), ConfigValue: c.config, PlanValue: types.StringUnknown(), Plan: plan}
			resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}

			versionFromManualVersion().PlanModifyString(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}
			if !resp.PlanValue.Equal(c.want) {
				t.Errorf("got %s, want %s", resp.PlanValue, c.want)
			}
		})
	}
}

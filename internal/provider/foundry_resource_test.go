package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func auth0Config(secret string, version int64) *FoundryAuth0Model {
	return &FoundryAuth0Model{
		Domain:                types.StringValue("tenant.eu.auth0.com"),
		ClientID:              types.StringValue("client"),
		ClientSecretWO:        types.StringValue(secret),
		ClientSecretWOVersion: types.Int64Value(version),
	}
}

func TestFoundryUpdateInputSendsSecretOnNewVersion(t *testing.T) {
	tests := []struct {
		name  string
		state *FoundryAuth0Model
		want  bool
	}{
		{"block newly managed", nil, true},
		{"version changed", auth0Config("", 1), true},
		{"version unchanged", auth0Config("", 2), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &FoundryResourceModel{Auth0Config: auth0Config("secret", 2)}
			plan := &FoundryResourceModel{Auth0Config: auth0Config("", 2)}
			plan.Auth0Config.ClientSecretWO = types.StringNull()

			input := foundryUpdateInput(plan, config, &FoundryResourceModel{Auth0Config: test.state})

			if got := input.Auth0ClientSecret != nil; got != test.want {
				t.Fatalf("secret sent = %v, want %v", got, test.want)
			}
			if test.want && *input.Auth0ClientSecret != "secret" {
				t.Fatalf("secret = %q", *input.Auth0ClientSecret)
			}
			if *input.Auth0Domain != "tenant.eu.auth0.com" {
				t.Fatalf("domain = %q", *input.Auth0Domain)
			}
		})
	}
}

func TestFoundryUpdateInputLeavesUnmanagedBlocksAlone(t *testing.T) {
	input := foundryUpdateInput(&FoundryResourceModel{}, &FoundryResourceModel{}, &FoundryResourceModel{})

	if input.Auth0Domain != nil || input.SentryEnabled != nil || input.SentryDSN != nil || input.LetsencryptEnabled != nil ||
		input.SessionIdleTimeoutHours != nil || input.SessionMaxAgeDays != nil {
		t.Fatalf("sent settings of a block left out: %+v", input)
	}
}

func TestFoundryUpdateInputClearsLeftOutFields(t *testing.T) {
	plan := &FoundryResourceModel{LetsencryptConfig: &FoundryLetsencryptModel{
		Enabled: types.BoolValue(false), Email: types.StringNull(), Staging: types.BoolValue(false),
	}}

	input := foundryUpdateInput(plan, plan, &FoundryResourceModel{})

	if input.LetsencryptEmail == nil || *input.LetsencryptEmail != "" {
		t.Fatalf("email = %v, want it cleared", input.LetsencryptEmail)
	}
}

func TestFoundryUpdateInputSendsSessionSettings(t *testing.T) {
	plan := &FoundryResourceModel{SessionConfig: &FoundrySessionModel{
		IdleTimeoutHours: types.Int64Value(8), MaxAgeDays: types.Int64Value(7),
	}}

	input := foundryUpdateInput(plan, plan, &FoundryResourceModel{})

	if input.SessionIdleTimeoutHours == nil || *input.SessionIdleTimeoutHours != 8 ||
		input.SessionMaxAgeDays == nil || *input.SessionMaxAgeDays != 7 {
		t.Fatalf("got %v and %v, want 8 hours and 7 days", input.SessionIdleTimeoutHours, input.SessionMaxAgeDays)
	}
}

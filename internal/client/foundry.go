package client

import (
	"context"
)

type FoundryRequest struct {
	client *Client
}

type FoundryResult struct {
	Foundry   *Foundry   `json:"foundry"`
	Operation *Operation `json:"operation,omitempty"`
}

type Foundry struct {
	ID                 int64  `json:"id"`
	Status             string `json:"status"`
	HATopology         string `json:"ha_topology"`
	DesiredHATopology  string `json:"desired_ha_topology"`
	Revision           int64  `json:"revision"`
	Version            string `json:"version"`
	ImageRepository    string `json:"image_repository"`
	ImageTag           string `json:"image_tag"`
	RailsLogLevel      string `json:"rails_log_level"`
	IdentityProvider   string `json:"identity_provider"`
	Auth0Domain        string `json:"auth0_domain"`
	Auth0ClientID      string `json:"auth0_client_id"`
	SentryEnabled      bool   `json:"sentry_enabled"`
	SentryEnvironment  string `json:"sentry_environment"`
	LetsencryptEnabled bool   `json:"letsencrypt_enabled"`
	LetsencryptEmail   string `json:"letsencrypt_email"`
	LetsencryptStaging bool   `json:"letsencrypt_staging"`

	SessionIdleTimeoutHours int64 `json:"session_idle_timeout_hours"`
	SessionMaxAgeDays       int64 `json:"session_max_age_days"`
}

// A nil field keeps the setting, an empty string clears it. An empty secret keeps the stored one.
type FoundryUpdateInput struct {
	Version            *string `json:"version,omitempty"`
	RailsLogLevel      *string `json:"rails_log_level,omitempty"`
	IdentityProvider   *string `json:"identity_provider,omitempty"`
	Auth0Domain        *string `json:"auth0_domain,omitempty"`
	Auth0ClientID      *string `json:"auth0_client_id,omitempty"`
	Auth0ClientSecret  *string `json:"auth0_client_secret,omitempty"`
	SentryEnabled      *bool   `json:"sentry_enabled,omitempty"`
	SentryDSN          *string `json:"sentry_dsn,omitempty"`
	SentryEnvironment  *string `json:"sentry_environment,omitempty"`
	LetsencryptEnabled *bool   `json:"letsencrypt_enabled,omitempty"`
	LetsencryptEmail   *string `json:"letsencrypt_email,omitempty"`
	LetsencryptStaging *bool   `json:"letsencrypt_staging,omitempty"`

	SessionIdleTimeoutHours *int64 `json:"session_idle_timeout_hours,omitempty"`
	SessionMaxAgeDays       *int64 `json:"session_max_age_days,omitempty"`
}

func (c *Client) Foundry() *FoundryRequest {
	return &FoundryRequest{client: c}
}

func (fr *FoundryRequest) Get(ctx context.Context) (*FoundryResult, *Error) {
	return fr.do(ctx, fr.client.Get, &ClientRequest{Path: "foundry", Result: &FoundryResult{}})
}

func (fr *FoundryRequest) Update(ctx context.Context, input *FoundryUpdateInput) (*FoundryResult, *Error) {
	return fr.do(ctx, fr.client.Put, &ClientRequest{Path: "foundry", Result: &FoundryResult{}, Body: input})
}

func (fr *FoundryRequest) do(ctx context.Context, send func(context.Context, *ClientRequest) (any, *Error), clientRequest *ClientRequest) (*FoundryResult, *Error) {
	result, err := send(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	foundryResult, ok := result.(*FoundryResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return foundryResult, nil
}

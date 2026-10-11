package client

import "context"

type FoundryVersion struct {
	Name     string   `json:"name"`
	Channels []string `json:"channels"`
	EOLDate  string   `json:"eol_date"`
	Revoked  bool     `json:"revoked"`
}

type NestVersion struct {
	Name              string `json:"name"`
	KubernetesVersion string `json:"kubernetes_version"`
	EOLDate           string `json:"eol_date"`
	Revoked           bool   `json:"revoked"`
}

type FoundryVersionsResult struct {
	FoundryVersions []*FoundryVersion `json:"foundry_versions"`
}

type NestVersionsResult struct {
	NestVersions []*NestVersion `json:"nest_versions"`
}

// Newest first.
func (c *Client) FoundryVersions(ctx context.Context) (*FoundryVersionsResult, *Error) {
	result, err := c.Get(ctx, &ClientRequest{Path: "foundry_versions", Result: &FoundryVersionsResult{}})
	if err != nil {
		return nil, err
	}

	versions, ok := result.(*FoundryVersionsResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}
	return versions, nil
}

// Newest first.
func (c *Client) NestVersions(ctx context.Context) (*NestVersionsResult, *Error) {
	result, err := c.Get(ctx, &ClientRequest{Path: "nest_versions", Result: &NestVersionsResult{}})
	if err != nil {
		return nil, err
	}

	versions, ok := result.(*NestVersionsResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}
	return versions, nil
}

package client

import (
	"context"
	"fmt"
)

type NetworkProfileRequest struct {
	client *Client
}

type NetworkProfileResult struct {
	NetworkProfile *NetworkProfile `json:"network_profile"`
	Operation      *Operation      `json:"operation,omitempty"`
}

type NetworkProfile struct {
	ID      int64    `json:"id"`
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Uplinks []Uplink `json:"uplinks"`
}

type NetworkProfileCreateInput struct {
	Name    string   `json:"name"`
	Uplinks []Uplink `json:"uplinks"`
}

// Uplink is one group of interfaces on a machine. Mode says how many to expect
// — auto, single or bond — and Identifier what Interfaces are matched against,
// kernel_name or mac_address.
type Uplink struct {
	Name         string        `json:"name"`
	Mode         string        `json:"mode"`
	Identifier   string        `json:"identifier,omitempty"`
	Interfaces   []string      `json:"interfaces,omitempty"`
	LACP         bool          `json:"lacp"`
	HostNetworks []HostNetwork `json:"host_networks"`
}

// HostNetwork attaches an uplink to a subnet. Exactly one across the profile is
// primary, which decides the default route, DNS and NTP.
type HostNetwork struct {
	SubnetID   int64 `json:"subnet_id"`
	VLANTagged bool  `json:"vlan_tagged"`
	Primary    bool  `json:"primary"`
}

func (c *Client) NetworkProfile() *NetworkProfileRequest {
	return &NetworkProfileRequest{
		client: c,
	}
}

func (mr *NetworkProfileRequest) Get(ctx context.Context, id int64) (*NetworkProfileResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "network_profiles", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &NetworkProfileResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)

	if err != nil {
		return nil, err
	}

	profileResult, ok := result.(*NetworkProfileResult)

	if !ok {
		return nil, &ErrorTypeAssert
	}

	return profileResult, nil
}

func (mr *NetworkProfileRequest) Create(ctx context.Context, input *NetworkProfileCreateInput) (*NetworkProfileResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "network_profiles",
		Result: &NetworkProfileResult{},
		Body:   map[string]any{"network_profile": input},
	}

	result, err := mr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	profileResult, ok := result.(*NetworkProfileResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return profileResult, nil
}

func (mr *NetworkProfileRequest) Delete(ctx context.Context, id int64) (*NetworkProfileResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "network_profiles", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &NetworkProfileResult{},
	}

	result, err := mr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	profileResult, ok := result.(*NetworkProfileResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return profileResult, nil
}

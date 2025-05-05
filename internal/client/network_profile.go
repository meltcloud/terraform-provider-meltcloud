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
	VLANs   []VLAN   `json:"vlans"`
	Bridges []Bridge `json:"bridges"`
	Bonds   []Bond   `json:"bonds"`
}

type NetworkProfileCreateInput struct {
	Name    string   `json:"name"`
	VLANs   []VLAN   `json:"vlans"`
	Bridges []Bridge `json:"bridges"`
	Bonds   []Bond   `json:"bonds"`
}

type VLAN struct {
	VLAN      int64  `json:"vlan"`
	DHCP      bool   `json:"dhcp"`
	Interface string `json:"interface"`
}

type Bridge struct {
	Name      string `json:"name"`
	Interface string `json:"interface"`
	DHCP      bool   `json:"dhcp"`
}

type Bond struct {
	Name       string `json:"name"`
	Interfaces string `json:"interfaces"`
	Kind       string `json:"kind"`
	DHCP       bool   `json:"dhcp"`
}

type NetworkProfileUpdateInput struct {
	Name    string   `json:"name"`
	VLANs   []VLAN   `json:"vlans"`
	Bridges []Bridge `json:"bridges"`
	Bonds   []Bond   `json:"bonds"`
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
		Body:   input,
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

func (mr *NetworkProfileRequest) Update(ctx context.Context, id int64, input *NetworkProfileUpdateInput) (*NetworkProfileResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "network_profiles", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &NetworkProfileResult{},
		Body:   input,
	}

	result, err := mr.client.Put(ctx, clientRequest)
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

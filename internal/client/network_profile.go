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
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Links  []Link `json:"links"`
}

type NetworkProfileCreateInput struct {
	Name  string `json:"name"`
	Links []Link `json:"links"`
}

type Link struct {
	Name           string   `json:"name"`
	Interfaces     []string `json:"interfaces"`
	VLANs          []int64  `json:"vlans"`
	HostNetworking bool     `json:"host_networking"`
	LACP           bool     `json:"lacp"`
	NativeVLAN     bool     `json:"native_vlan"`
}

type NetworkProfileUpdateInput struct {
	Name  string `json:"name"`
	Links []Link `json:"links"`
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

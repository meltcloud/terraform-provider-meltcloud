package client

import (
	"context"
	"fmt"
)

type SubnetRequest struct {
	client *Client
}

type SubnetResult struct {
	Subnet    *Subnet    `json:"subnet"`
	Operation *Operation `json:"operation,omitempty"`
}

type Subnet struct {
	ID         int64         `json:"id"`
	NetworkID  int64         `json:"network_id"`
	Name       string        `json:"name"`
	VLAN       *int64        `json:"vlan"`
	Addressing string        `json:"addressing"`
	IPPoolID   *int64        `json:"ip_pool_id"`
	Gateway    *string       `json:"gateway"`
	DNS        []string      `json:"dns"`
	NTP        []string      `json:"ntp"`
	Domains    []string      `json:"domains"`
	MTU        *int64        `json:"mtu"`
	Routes     []SubnetRoute `json:"routes"`
}

type SubnetRoute struct {
	Destination string `json:"destination"`
	Via         string `json:"via"`
	Metric      *int64 `json:"metric"`
}

type SubnetCreateInput struct {
	Name       string        `json:"name"`
	VLAN       *int64        `json:"vlan,omitempty"`
	Addressing string        `json:"addressing"`
	IPPoolID   *int64        `json:"ip_pool_id,omitempty"`
	Gateway    *string       `json:"gateway,omitempty"`
	DNS        []string      `json:"dns,omitempty"`
	NTP        []string      `json:"ntp,omitempty"`
	Domains    []string      `json:"domains,omitempty"`
	MTU        *int64        `json:"mtu,omitempty"`
	Routes     []SubnetRoute `json:"routes,omitempty"`
}

func (c *Client) Subnet() *SubnetRequest {
	return &SubnetRequest{client: c}
}

func subnetPath(networkID int64, parts ...any) string {
	path := fmt.Sprintf("%s/%d/%s", "networks", networkID, "subnets")
	for _, part := range parts {
		path = fmt.Sprintf("%s/%v", path, part)
	}
	return path
}

func (sr *SubnetRequest) Get(ctx context.Context, networkID int64, id int64) (*SubnetResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   subnetPath(networkID, id),
		Result: &SubnetResult{},
	}

	result, err := sr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	subnetResult, ok := result.(*SubnetResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return subnetResult, nil
}

func (sr *SubnetRequest) Create(ctx context.Context, networkID int64, input *SubnetCreateInput) (*SubnetResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   subnetPath(networkID),
		Result: &SubnetResult{},
		Body:   map[string]any{"subnet": input},
	}

	result, err := sr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	subnetResult, ok := result.(*SubnetResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return subnetResult, nil
}

func (sr *SubnetRequest) Delete(ctx context.Context, networkID int64, id int64) (*SubnetResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   subnetPath(networkID, id),
		Result: &SubnetResult{},
	}

	result, err := sr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	subnetResult, ok := result.(*SubnetResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return subnetResult, nil
}

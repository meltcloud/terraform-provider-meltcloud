package client

import (
	"context"
	"fmt"
)

type MachinePoolRequest struct {
	client *Client
}

type MachinePoolResult struct {
	MachinePool *MachinePool `json:"machine_pool"`
	Operation   *Operation   `json:"operation,omitempty"`
}

type MachinePool struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	UserVersion  string `json:"user_version"`
	PatchVersion string `json:"patch_version"`
}

type MachinePoolCreateInput struct {
	Name              string `json:"name"`
	UserVersion       string `json:"user_version"`
	PrimaryDiskDevice string `json:"primary_disk_device"`
}

type MachinePoolUpdateInput struct {
	Name              string `json:"name"`
	UserVersion       string `json:"user_version"`
	PrimaryDiskDevice string `json:"primary_disk_device"`
}

func (c *Client) MachinePool() *MachinePoolRequest {
	return &MachinePoolRequest{
		client: c,
	}
}

func (mr *MachinePoolRequest) Get(ctx context.Context, clusterId int64, id int64) (*MachinePoolResult, *Error) {
	subPath := fmt.Sprintf("%s/%d/%s/%d", "clusters", clusterId, "machine_pools", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &MachinePoolResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	machineResult, ok := result.(*MachinePoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return machineResult, nil
}

func (mr *MachinePoolRequest) Create(ctx context.Context, clusterId int64, input *MachinePoolCreateInput) (*MachinePoolResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   fmt.Sprintf("%s/%d/%s", "clusters", clusterId, "machine_pools"),
		Result: &MachinePoolResult{},
		Body:   input,
	}

	result, err := mr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	machineResult, ok := result.(*MachinePoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return machineResult, nil
}

func (mr *MachinePoolRequest) Update(ctx context.Context, clusterId int64, id int64, input *MachinePoolUpdateInput) (*MachinePoolResult, *Error) {
	subPath := fmt.Sprintf("%s/%d/%s/%d", "clusters", clusterId, "machine_pools", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &MachinePoolResult{},
		Body:   input,
	}

	result, err := mr.client.Put(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	machineResult, ok := result.(*MachinePoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return machineResult, nil
}

func (mr *MachinePoolRequest) Delete(ctx context.Context, clusterId int64, id int64) (*MachinePoolResult, *Error) {
	subPath := fmt.Sprintf("%s/%d/%s/%d", "clusters", clusterId, "machine_pools", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &MachinePoolResult{},
	}

	result, err := mr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	machineResult, ok := result.(*MachinePoolResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return machineResult, nil
}

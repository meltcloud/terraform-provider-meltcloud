package client

import (
	"context"
	"fmt"
	"github.com/google/uuid"
)

type MachineRequest struct {
	client *Client
}

type MachineResult struct {
	Machine   *Machine   `json:"machine"`
	Operation *Operation `json:"operation,omitempty"`
}

type MachinesResult struct {
	Machines []*Machine `json:"machines"`
}

type Machine struct {
	ID            int64     `json:"id"`
	UUID          uuid.UUID `json:"uuid"`
	Name          string    `json:"name,omitempty"`
	Status        string    `json:"status"`
	MachinePoolID int64     `json:"machine_pool_id,omitempty"`
	Labels        []Label   `json:"labels,omitempty"`
}

type MachineCreateInput struct {
	UUID          uuid.UUID `json:"uuid"`
	Name          string    `json:"name,omitempty"`
	MachinePoolID int64     `json:"machine_pool_id,omitempty"`
	Labels        []Label   `json:"labels,omitempty"`
}

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MachineUpdateInput struct {
	Name          string  `json:"name,omitempty"`
	MachinePoolID int64   `json:"machine_pool_id,omitempty"`
	Labels        []Label `json:"labels,omitempty"`
}

func (c *Client) Machine() *MachineRequest {
	return &MachineRequest{
		client: c,
	}
}

func (mr *MachineRequest) List(ctx context.Context) (*MachinesResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "machines",
		Result: &MachinesResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	machinesResult, ok := result.(*MachinesResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return machinesResult, nil
}

func (mr *MachineRequest) Get(ctx context.Context, id int64) (*MachineResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "machines", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &MachineResult{},
	}

	result, err := mr.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	machineResult, ok := result.(*MachineResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return machineResult, nil
}

func (mr *MachineRequest) Create(ctx context.Context, input *MachineCreateInput) (*MachineResult, *Error) {
	clientRequest := &ClientRequest{
		Path:   "machines",
		Result: &MachineResult{},
		Body:   input,
	}

	result, err := mr.client.Post(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	machineResult, ok := result.(*MachineResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return machineResult, nil
}

func (mr *MachineRequest) Update(ctx context.Context, id int64, input *MachineUpdateInput) (*MachineResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "machines", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &MachineResult{},
		Body:   input,
	}

	result, err := mr.client.Put(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	machineResult, ok := result.(*MachineResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return machineResult, nil
}

func (mr *MachineRequest) Delete(ctx context.Context, id int64) (*MachineResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "machines", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &MachineResult{},
	}

	result, err := mr.client.Delete(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	machineResult, ok := result.(*MachineResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return machineResult, nil
}

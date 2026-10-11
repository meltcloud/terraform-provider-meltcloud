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
	ID                    int64            `json:"id"`
	UUID                  uuid.UUID        `json:"uuid"`
	Name                  string           `json:"name,omitempty"`
	Status                string           `json:"status"`
	MachinePoolID         *int64           `json:"machine_pool_id"`
	NetworkProfileID      int64            `json:"network_profile_id"`
	DepotNetworkProfileID int64            `json:"depot_network_profile_id"`
	ZoneID                *int64           `json:"zone_id"`
	Labels                []Label          `json:"labels,omitempty"`
	Revision              *MachineRevision `json:"machine_revision,omitempty"`
}

type MachineRevision struct {
	IPAddresses []IPAddress `json:"ip_addresses"`
}

// An address the Machine holds in an IP Pool. Source is "operator" for a pinned one and
// "ipam" for one the pool allocated.
type IPAddress struct {
	IPPoolID int64  `json:"ip_pool_id"`
	Address  string `json:"address"`
	Source   string `json:"source,omitempty"`
}

type MachineCreateInput struct {
	UUID                  uuid.UUID   `json:"uuid"`
	Name                  string      `json:"name,omitempty"`
	MachinePoolID         int64       `json:"machine_pool_id,omitempty"`
	NetworkProfileID      int64       `json:"network_profile_id"`
	DepotNetworkProfileID int64       `json:"depot_network_profile_id"`
	ZoneID                *int64      `json:"zone_id,omitempty"`
	Labels                []Label     `json:"labels,omitempty"`
	IPAddresses           []IPAddress `json:"ip_addresses,omitempty"`
}

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Sends every field, so an empty list or a null zone_id clears what the Machine has.
type MachineUpdateInput struct {
	Name                  string      `json:"name,omitempty"`
	MachinePoolID         int64       `json:"machine_pool_id,omitempty"`
	NetworkProfileID      int64       `json:"network_profile_id"`
	DepotNetworkProfileID int64       `json:"depot_network_profile_id"`
	ZoneID                *int64      `json:"zone_id"`
	Labels                []Label     `json:"labels"`
	IPAddresses           []IPAddress `json:"ip_addresses"`
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

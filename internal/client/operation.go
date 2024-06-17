package client

import (
	"context"
	"fmt"
	"time"
)

type OperationRequest struct {
	client *Client
}

type OperationResult struct {
	Operation *Operation `json:"operation"`
}

type Operation struct {
	ID     int64           `json:"id"`
	Status OperationStatus `json:"status"`
	Action string          `json:"action"`
}

type OperationStatus string

const (
	OperationStatusPending   OperationStatus = "pending"
	OperationStatusRunning   OperationStatus = "running"
	OperationStatusSucceeded OperationStatus = "succeeded"
	OperationStatusFailed    OperationStatus = "failed"
)

func (c *Client) Operation() *OperationRequest {
	return &OperationRequest{
		client: c,
	}
}

func (or *OperationRequest) Get(ctx context.Context, id int64) (*OperationResult, *Error) {
	subPath := fmt.Sprintf("%s/%d", "operations", id)
	clientRequest := &ClientRequest{
		Path:   subPath,
		Result: &OperationResult{},
	}

	result, err := or.client.Get(ctx, clientRequest)
	if err != nil {
		return nil, err
	}

	operationResult, ok := result.(*OperationResult)
	if !ok {
		return nil, &ErrorTypeAssert
	}

	return operationResult, nil
}

func (or *OperationRequest) PollUntilDone(ctx context.Context, id int64) (*OperationResult, *Error) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			result, err := or.Get(ctx, id)
			if err != nil {
				return nil, err
			}
			if result.Operation.Status == OperationStatusSucceeded || result.Operation.Status == OperationStatusFailed {
				return result, nil
			}
		case <-ctx.Done():
			return nil, &Error{Err: ctx.Err()}
		}
	}
}

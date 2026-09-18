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
	OperationStatusPending                 OperationStatus = "pending"
	OperationStatusRunning                 OperationStatus = "running"
	OperationStatusCancelling              OperationStatus = "cancelling"
	OperationStatusSucceeded               OperationStatus = "succeeded"
	OperationStatusFailed                  OperationStatus = "failed"
	OperationStatusCancelled               OperationStatus = "cancelled"
	OperationStatusInternalError           OperationStatus = "internal_error"
	OperationStatusCancellingInternalError OperationStatus = "cancelling_internal_error"
	OperationStatusRetryableInternalError  OperationStatus = "retryable_internal_error"
)

// The statuses foundry does not move an operation out of.
func (s OperationStatus) Finished() bool {
	switch s {
	case OperationStatusSucceeded, OperationStatusFailed, OperationStatusCancelled,
		OperationStatusInternalError, OperationStatusRetryableInternalError:
		return true
	}
	return false
}

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
			if result.Operation.Status == OperationStatusSucceeded {
				return result, nil
			}
			if result.Operation.Status.Finished() {
				consoleURL := fmt.Sprintf("%s/ui/orgs/%s/operations/%d", or.client.Endpoint, or.client.Organization, id)
				return result, &Error{Err: fmt.Errorf("operation %d %s. check the console for more information: %s", result.Operation.ID, result.Operation.Status, consoleURL)}
			}
		case <-ctx.Done():
			return nil, &Error{Err: ctx.Err()}
		}
	}
}

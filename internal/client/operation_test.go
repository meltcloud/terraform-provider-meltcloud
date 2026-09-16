package client

import "testing"

func TestOperationStatusFinished(t *testing.T) {
	finished := []OperationStatus{
		OperationStatusSucceeded,
		OperationStatusFailed,
		OperationStatusCancelled,
		OperationStatusInternalError,
		OperationStatusRetryableInternalError,
	}
	for _, status := range finished {
		if !status.Finished() {
			t.Errorf("%s should be finished", status)
		}
	}

	unfinished := []OperationStatus{
		OperationStatusPending,
		OperationStatusRunning,
		OperationStatusCancelling,
		OperationStatusCancellingInternalError,
	}
	for _, status := range unfinished {
		if status.Finished() {
			t.Errorf("%s should not be finished", status)
		}
	}
}

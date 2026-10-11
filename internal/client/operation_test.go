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

func TestUnavailable(t *testing.T) {
	for _, status := range []int{0, 502, 503, 504} {
		if !unavailable(&Error{HTTPStatusCode: status}) {
			t.Errorf("status %d should count as the API restarting", status)
		}
	}
	for _, status := range []int{401, 404, 422, 500} {
		if unavailable(&Error{HTTPStatusCode: status}) {
			t.Errorf("status %d should end the poll", status)
		}
	}
}

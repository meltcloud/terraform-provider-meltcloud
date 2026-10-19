package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

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

// Answers 503 for the first unavailable polls, then reports the operation as succeeded.
func pollServer(t *testing.T, unavailable int32) *Client {
	var polls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if polls.Add(1) <= unavailable {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"operation":{"id":1,"status":"succeeded"}}`))
	}))
	t.Cleanup(server.Close)

	interval, limit := pollInterval, maxUnavailable
	pollInterval, maxUnavailable = time.Millisecond, 50*time.Millisecond
	t.Cleanup(func() { pollInterval, maxUnavailable = interval, limit })
	return New(server.URL, "org", "key", nil)
}

func TestPollUntilDoneWaitsOutTheAPIRestarting(t *testing.T) {
	result, err := pollServer(t, 3).Operation().PollUntilDone(context.Background(), 1)
	if err != nil || result.Operation.Status != OperationStatusSucceeded {
		t.Fatalf("got %+v, %v; want the operation to succeed", result, err)
	}
}

func TestPollUntilDoneGivesUpOnAnAPIThatStaysDown(t *testing.T) {
	_, err := pollServer(t, 1_000_000).Operation().PollUntilDone(context.Background(), 1)
	if err == nil || err.HTTPStatusCode != http.StatusServiceUnavailable {
		t.Fatalf("got %v; want the 503 once the API stayed down too long", err)
	}
}

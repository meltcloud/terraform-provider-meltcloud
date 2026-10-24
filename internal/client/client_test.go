package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRedirectsAreNotFollowed(t *testing.T) {
	var leaked atomic.Bool
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leaked.Store(r.Header.Get("X-Meltcloud-API-Key") != "")
	}))
	t.Cleanup(other.Close)
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL, http.StatusTemporaryRedirect)
	}))
	t.Cleanup(api.Close)

	_, err := New(api.URL, "org", "key", nil).Get(context.Background(), &ClientRequest{Path: "clusters"})
	if err == nil {
		t.Fatal("got no error; want the redirect refused")
	}
	if leaked.Load() {
		t.Fatal("the API key reached the redirect target")
	}
}

func TestRateLimitedRequestsAreRetried(t *testing.T) {
	var requests atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(api.Close)

	start := time.Now()
	_, err := New(api.URL, "org", "key", nil).Post(context.Background(), &ClientRequest{Path: "clusters", Body: map[string]string{}})
	if err != nil {
		t.Fatalf("got %v; want the request to succeed after the 429", err)
	}
	if requests.Load() != 2 || time.Since(start) < time.Second {
		t.Fatalf("got %d requests in %s; want a retry after Retry-After", requests.Load(), time.Since(start))
	}
}

func TestFailedRequestsAreNotRetried(t *testing.T) {
	var requests atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(api.Close)

	_, _ = New(api.URL, "org", "key", nil).Post(context.Background(), &ClientRequest{Path: "clusters", Body: map[string]string{}})
	if requests.Load() != 1 {
		t.Fatalf("got %d requests; want a create sent once", requests.Load())
	}
}

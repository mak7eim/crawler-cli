package fetcher

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchDoesNotFollowRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/target", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	fetcher := New(time.Second)
	body, status, _, err := fetcher.Fetch(context.Background(), server.URL+"/start")
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	defer body.Close()
	if status != http.StatusFound {
		t.Fatalf("status = %d, want %d", status, http.StatusFound)
	}
	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("redirect response body is empty")
	}
}

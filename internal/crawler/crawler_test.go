package crawler

import (
	"context"
	"crawler-cli/internal/fetcher"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"crawler-cli/internal/logger"

	"go.uber.org/goleak"
)

type blockingFetcher struct {
	started chan struct{}
}

func (f *blockingFetcher) Fetch(
	ctx context.Context,
	_ string,
) (io.ReadCloser, int, string, error) {
	close(f.started)

	<-ctx.Done()
	return nil, 0, "", ctx.Err()
}

func TestRunStopsWorkersAfterOverallTimeout(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	log, err := logger.New(t.TempDir() + "/crawler.log")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()

	f := &blockingFetcher{
		started: make(chan struct{}),
	}
	craw := New(0, f, log)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	runDone := make(chan struct{})
	go func() {
		craw.Run(ctx, []string{"https://example.test"})
		close(runDone)
	}()

	select {
	case <-f.started:
	case <-time.After(time.Second):
		t.Fatal("fetch task not start")
	}

	select {
	case <-runDone:
	case <-time.After(time.Second):
		t.Fatal("Run not return after context timeout")
	}

	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("context error = %v, want DeadlineExceeded", ctx.Err())
	}
}

func TestRunProcessesMoreThanWorkerBatch(t *testing.T) {
	const children = 25

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if r.URL.Path == "/" {
			fmt.Fprint(w, "<title>root</title>")
			for i := 0; i < children; i++ {
				fmt.Fprintf(w, `<a href="/child/%d">child</a>`, i)
			}
			return
		}
		fmt.Fprintf(w, "<title>%s</title>", r.URL.Path)
	}))
	defer server.Close()

	log, err := logger.New(t.TempDir() + "/crawler.log")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()

	crawler := New(1, fetcher.New(time.Second), log)
	tree := crawler.Run(context.Background(), []string{server.URL + "/"})

	if len(tree) != 1 {
		t.Fatalf("top-level nodes = %d, want 1", len(tree))
	}
	if len(tree[0].Links) != children {
		t.Fatalf("processed child pages = %d, want %d", len(tree[0].Links), children)
	}
}

func TestIsHTML(t *testing.T) {
	tests := []struct {
		contentType string
		want        bool
	}{
		{contentType: "text/html; charset=utf-8", want: true},
		{contentType: "TEXT/HTML", want: true},
		{contentType: "text/html-invalid", want: false},
		{contentType: "application/json", want: false},
	}

	for _, test := range tests {
		if got := isHTML(test.contentType); got != test.want {
			t.Errorf("isHTML(%q) = %t, want %t", test.contentType, got, test.want)
		}
	}
}

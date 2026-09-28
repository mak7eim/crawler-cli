package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"crawler-cli/internal/logger"
)

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

	crawler := New(1, time.Second, log)
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

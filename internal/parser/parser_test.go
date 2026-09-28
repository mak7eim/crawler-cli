package parser

import (
	"net/url"
	"testing"
)

func TestParseExtractsTitleAndSameHostLinks(t *testing.T) {
	base, err := url.Parse("https://example.com/docs/index.html")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Parse([]byte(`<html><head><title> A page </title></head><body>
		<a href="/about">About</a><a href="https://other.example/about">Other</a>
		<a href="/about">Duplicate</a></body></html>`), base)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.Title != "A page" {
		t.Fatalf("title = %q, want %q", result.Title, "A page")
	}
	if len(result.Links) != 1 || result.Links[0] != "https://example.com/about" {
		t.Fatalf("links = %#v", result.Links)
	}
}

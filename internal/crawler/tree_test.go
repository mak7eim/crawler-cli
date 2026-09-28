package crawler

import (
	"encoding/json"
	"testing"
)

func TestBuildTreeSkipsErrorsAndBreaksCycles(t *testing.T) {
	results := []ResultTask{
		{Task: Task{URL: "https://example.com/a", Depth: 0}, Links: []string{"https://example.com/b"}},
		{Task: Task{URL: "https://example.com/b", Depth: 1, ParentURL: "https://example.com/a"}, Links: []string{"https://example.com/a"}},
		{Task: Task{URL: "https://example.com/bad"}, Err: errTest{}},
	}

	tree := buildTree(results, []string{"https://example.com/a"})
	data, err := json.Marshal(tree)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if string(data) == "" || len(tree) != 1 || len(tree[0].Links) != 1 {
		t.Fatalf("unexpected tree: %s", data)
	}
	if tree[0].Links[0].Resource != "https://example.com/b" {
		t.Fatalf("unexpected child: %#v", tree[0].Links[0])
	}
}

func TestBuildTreeUsesDiscoveryParent(t *testing.T) {
	results := []ResultTask{
		{Task: Task{URL: "https://example.com/a", Depth: 0}},
		{Task: Task{URL: "https://example.com/b", Depth: 1, ParentURL: "https://example.com/a"}},
		{Task: Task{URL: "https://example.com/c", Depth: 2, ParentURL: "https://example.com/b"}},
	}

	tree := buildTree(results, []string{"https://example.com/a"})
	if len(tree) != 1 || len(tree[0].Links) != 1 || len(tree[0].Links[0].Links) != 1 {
		t.Fatalf("unexpected tree shape: %#v", tree)
	}
	if tree[0].Links[0].Links[0].Resource != "https://example.com/c" {
		t.Fatalf("c has wrong parent: %#v", tree[0].Links[0].Links)
	}
}

func TestBuildTreeReturnsEmptySlice(t *testing.T) {
	tree := buildTree([]ResultTask{{Task: Task{URL: "https://example.com"}, Err: errTest{}}}, []string{"https://example.com"})
	if tree == nil || len(tree) != 0 {
		t.Fatalf("tree = %#v, want empty non-nil slice", tree)
	}

	data, err := json.Marshal(tree)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if string(data) != "[]" {
		t.Fatalf("JSON = %s, want []", data)
	}
}

type errTest struct{}

func (errTest) Error() string { return "test error" }

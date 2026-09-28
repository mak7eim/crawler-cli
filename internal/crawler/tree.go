package crawler

import (
	"encoding/json"
	"fmt"
	"os"
)

type PageNode struct {
	Resource string      `json:"resource"`
	Title    string      `json:"title"`
	Links    []*PageNode `json:"links"`
}

func buildTree(results []ResultTask, rootURLs []string) []*PageNode {
	nodes := make(map[string]*PageNode, len(results))
	depthByURL := make(map[string]int, len(results))
	for _, r := range results {
		if r.Err != nil {
			continue
		}
		depthByURL[r.Task.URL] = r.Task.Depth
		nodes[r.Task.URL] = &PageNode{
			Resource: r.Task.URL,
			Title:    r.Title,
			Links:    []*PageNode{},
		}
	}

	roots := make([]*PageNode, 0, len(rootURLs))
	rootAdded := make(map[string]struct{}, len(rootURLs))
	for _, u := range rootURLs {
		if _, alreadyAdded := rootAdded[u]; alreadyAdded {
			continue
		}
		if n, ok := nodes[u]; ok {
			roots = append(roots, n)
			rootAdded[u] = struct{}{}
		}
	}

	for _, r := range results {
		if r.Err != nil || r.Task.ParentURL == "" {
			continue
		}
		parent, parentExists := nodes[r.Task.ParentURL]
		child, childExists := nodes[r.Task.URL]
		if !parentExists || !childExists || r.Task.ParentURL == r.Task.URL {
			continue
		}
		if depthByURL[r.Task.URL] != depthByURL[r.Task.ParentURL]+1 {
			continue
		}
		parent.Links = append(parent.Links, child)
	}

	return roots
}

func WriteJSON(path string, tree []*PageNode) error {
	data, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tree: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

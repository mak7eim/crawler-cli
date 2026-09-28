package parser

import (
	"bytes"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type Result struct {
	Title string
	Links []string
}

func Parse(body []byte, baseURL *url.URL) (Result, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}

	var raw []string
	walk(doc, &raw)

	result := Result{}
	result.Title = extractTitle(doc)

	seen := make(map[string]struct{})
	for _, href := range raw {
		ref, err := url.Parse(href)
		if err != nil {
			continue
		}
		abs := baseURL.ResolveReference(ref)
		if abs.Scheme != "http" && abs.Scheme != "https" {
			continue
		}
		if !strings.EqualFold(abs.Hostname(), baseURL.Hostname()) {
			continue
		}
		s := abs.String()
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		result.Links = append(result.Links, s)
	}

	return result, nil
}

func walk(node *html.Node, links *[]string) {
	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				*links = append(*links, attr.Val)
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walk(child, links)
	}
}

func extractTitle(node *html.Node) string {
	if node.Type == html.ElementNode && node.Data == "title" {
		var text strings.Builder
		collectText(node, &text)
		return strings.TrimSpace(text.String())
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if t := extractTitle(child); t != "" {
			return t
		}
	}
	return ""
}

func collectText(node *html.Node, text *strings.Builder) {
	if node.Type == html.TextNode {
		text.WriteString(node.Data)
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		collectText(child, text)
	}
}

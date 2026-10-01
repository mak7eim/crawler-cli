package crawler

import (
	"context"
	"crawler-cli/internal/logger"
	"crawler-cli/internal/parser"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const maxWorkers = 10

type Fetcher interface {
	Fetch(ctx context.Context, rawURL string) (io.ReadCloser, int, string, error)
}

type Crawler struct {
	depth      int
	maxWorkers int
	fetcher    Fetcher
	logger     *logger.Logger
}

func New(depth int, pageFetcher Fetcher, log *logger.Logger) *Crawler {
	return &Crawler{
		depth:      depth,
		maxWorkers: maxWorkers,
		fetcher:    pageFetcher,
		logger:     log,
	}
}

func (c *Crawler) Run(ctx context.Context, urls []string) []*PageNode {
	tasks := make(chan Task)
	results := make(chan ResultTask)

	visited := make(map[string]struct{}, len(urls))

	var collected []ResultTask

	var wg sync.WaitGroup
	wg.Add(c.maxWorkers)
	for i := 0; i < c.maxWorkers; i++ {
		go c.worker(ctx, tasks, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	pending := make([]Task, 0, len(urls))
	for _, u := range urls {
		if _, ok := visited[u]; ok {
			continue
		}
		visited[u] = struct{}{}
		pending = append(pending, Task{URL: u, Depth: 0})
	}

	inflight := 0
	for len(pending) > 0 || inflight > 0 {
		var taskChannel chan<- Task
		var nextTask Task
		if len(pending) > 0 {
			taskChannel = tasks
			nextTask = pending[0]
		}

		select {
		case <-ctx.Done():
			close(tasks)
			for r := range results {
				collected = append(collected, r)
			}
			return buildTree(collected, urls)

		case taskChannel <- nextTask:
			pending = pending[1:]
			inflight++

		case r, ok := <-results:
			if !ok {
				return buildTree(collected, urls)
			}
			inflight--
			collected = append(collected, r)

			if r.Task.Depth+1 > c.depth || r.Err != nil {
				continue
			}
			for _, link := range r.Links {
				if _, ok := visited[link]; ok {
					continue
				}
				visited[link] = struct{}{}
				pending = append(pending, Task{
					URL:       link,
					Depth:     r.Task.Depth + 1,
					ParentURL: r.Task.URL,
				})
			}
		}
	}

	close(tasks)
	for r := range results {
		collected = append(collected, r)
	}
	return buildTree(collected, urls)
}

func (c *Crawler) worker(
	ctx context.Context,
	tasks <-chan Task,
	results chan<- ResultTask,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-tasks:
			if !ok {
				return
			}
			res := c.process(ctx, task)
			results <- res
		}
	}
}

func (c *Crawler) process(ctx context.Context, task Task) ResultTask {
	res := ResultTask{Task: task}

	body, status, contentType, err := c.fetcher.Fetch(ctx, task.URL)
	res.Status = status

	if status != 0 {
		c.logger.Status(task.URL, status)
	}

	if err != nil {
		res.Err = fmt.Errorf("fetch %s: %w", task.URL, err)
		c.logger.Error("fetch url=%s err=%v", task.URL, err)
		return res
	}
	defer body.Close()
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		res.Err = fmt.Errorf("status %d for %s", status, task.URL)
		c.logger.Error("bad status url=%s status=%d", task.URL, status)
		return res
	}
	if !isHTML(contentType) {
		res.Err = fmt.Errorf("non-html content-type %q for %s", contentType, task.URL)
		c.logger.Error("non-html url=%s content-type=%q", task.URL, contentType)
		return res
	}

	data, err := io.ReadAll(body)
	if err != nil {
		res.Err = fmt.Errorf("read response body %s: %w", task.URL, err)
		c.logger.Error("read body url=%s err=%v", task.URL, err)
		return res
	}

	base, err := url.Parse(task.URL)
	if err != nil {
		res.Err = fmt.Errorf("parse base url %s: %w", task.URL, err)
		c.logger.Error("bad base url=%s err=%v", task.URL, err)
		return res
	}

	parsed, err := parser.Parse(data, base)
	if err != nil {
		res.Err = fmt.Errorf("parse html %s: %w", task.URL, err)
		c.logger.Error("parse html url=%s err=%v", task.URL, err)
		return res
	}

	res.Title = parsed.Title
	res.Links = parsed.Links
	return res
}

func isHTML(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && strings.EqualFold(mediaType, "text/html")
}

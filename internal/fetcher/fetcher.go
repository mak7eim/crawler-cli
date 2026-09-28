package fetcher

import (
	"context"
	"io"
	"net/http"
	"time"
)

type Fetcher struct {
	client *http.Client
}

func New(timeout time.Duration) *Fetcher {
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Fetcher{
		client: client,
	}
}

func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (io.ReadCloser, int, string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, "", err
	}

	response, err := f.client.Do(request)

	if err != nil {
		return nil, 0, "", err
	}

	contentType := response.Header.Get("Content-Type")
	return response.Body, response.StatusCode, contentType, nil
}

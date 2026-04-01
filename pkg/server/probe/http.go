package probe

import (
	"context"
	"fmt"
	"net/http"
)

// HTTPGet verifies that url responds with a 2xx or 3xx status code.
func HTTPGet(ctx context.Context, client *http.Client, url string) (err error) {
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close response body for %s: %w", url, closeErr)
		}
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}

	return nil
}

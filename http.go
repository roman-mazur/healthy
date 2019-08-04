package healthy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// An http client used for HTTP checks.
var httpClient = &http.Client{
	Timeout: 20 * time.Second,
}

// HttpCheck is a task that makes a GET HTTP request.
type HttpCheck struct {
	// What to check.
	Url string
	// Expected HTTP status code in the response.
	ExpectedStatusCode int
	// Request timeout.
	Timeout time.Duration
}

func (h *HttpCheck) Name() string {
	u, _ := url.Parse(h.Url)
	return fmt.Sprintf("HTTP check for %s", u.Host)
}

func (h *HttpCheck) Run(ctx context.Context) error {
	req, _ := http.NewRequest("GET", h.Url, nil)
	reqCtx := ctx
	if h.Timeout != 0 {
		reqCtx, _ = context.WithTimeout(ctx, h.Timeout)
	}
	req = req.WithContext(reqCtx)

	if resp, err := httpClient.Do(req); err == nil {
		if resp.StatusCode != h.ExpectedStatusCode {
			return fmt.Errorf("response code does not match: expected %d, got %d",
				h.ExpectedStatusCode, resp.StatusCode)
		}
	} else {
		return fmt.Errorf("issues performing a request; details: %s", err)
	}
	return nil
}

package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/nuvivo314/mcp-executor/internal/domain/port"
)

type httpClient struct {
	allowedHosts map[string]struct{}
	maxBodyBytes int64
	inner        *http.Client
}

// NewHTTPClient returns an HttpClient that enforces host allowlisting and body size cap.
// An empty allowedHosts slice means all hosts are permitted.
func NewHTTPClient(allowedHosts []string, maxBodyBytes int64) port.HttpClient {
	allowed := make(map[string]struct{}, len(allowedHosts))
	for _, h := range allowedHosts {
		allowed[h] = struct{}{}
	}
	return &httpClient{
		allowedHosts: allowed,
		maxBodyBytes: maxBodyBytes,
		inner:        &http.Client{},
	}
}

func (c *httpClient) Do(ctx context.Context, method, rawURL string, headers map[string]string, body string) (*port.HttpResponse, error) {
	if err := c.checkAllowlist(rawURL); err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.inner.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, c.maxBodyBytes)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	respHeaders := make(map[string]string, len(resp.Header))
	for k := range resp.Header {
		respHeaders[k] = resp.Header.Get(k)
	}

	return &port.HttpResponse{
		StatusCode: resp.StatusCode,
		Body:       string(raw),
		Headers:    respHeaders,
	}, nil
}

func (c *httpClient) checkAllowlist(rawURL string) error {
	if len(c.allowedHosts) == 0 {
		return nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parsing url %q: %w", rawURL, err)
	}
	host := u.Host // includes port if present
	// Also check without port
	hostOnly := u.Hostname()
	if _, ok := c.allowedHosts[host]; ok {
		return nil
	}
	if _, ok := c.allowedHosts[hostOnly]; ok {
		return nil
	}
	return fmt.Errorf("host %q is not in the allowlist", host)
}

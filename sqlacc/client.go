package sqlacc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"shwetaik-expense-management-api/configs"
)

const (
	defaultReadTimeout  = 30 * time.Second
	defaultWriteTimeout = 60 * time.Second
	defaultMaxAttempts  = 4
	backoffBase         = 500 * time.Millisecond
	backoffCap          = 8 * time.Second
)

type SQLAccClient struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

var (
	defaultOnce   sync.Once
	defaultClient *SQLAccClient
)

func Default() *SQLAccClient {
	defaultOnce.Do(func() {
		defaultClient = New(
			configs.Envs.SQLACC_API_ENDPOINT,
			configs.Envs.SQLACC_API_TOKEN,
		)
	})
	return defaultClient
}

func New(baseURL, token string) *SQLAccClient {
	return &SQLAccClient{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		},
	}
}

type RequestOpts struct {
	Timeout time.Duration
	MaxAttempts int
	RetryWrites bool
	Query url.Values
}

func (c *SQLAccClient) Get(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil, RequestOpts{Query: query})
}

func (c *SQLAccClient) Post(ctx context.Context, path string, body []byte) (*http.Response, error) {
	return c.Do(ctx, http.MethodPost, path, body, RequestOpts{})
}

func (c *SQLAccClient) Do(ctx context.Context, method, path string, body []byte, opts RequestOpts) (*http.Response, error) {
	isWrite := method != http.MethodGet && method != http.MethodHead

	if opts.Timeout == 0 {
		if isWrite {
			opts.Timeout = defaultWriteTimeout
		} else {
			opts.Timeout = defaultReadTimeout
		}
	}
	if opts.MaxAttempts == 0 {
		opts.MaxAttempts = defaultMaxAttempts
	}
	if isWrite && !opts.RetryWrites {
		opts.MaxAttempts = 1
	}

	fullURL := c.BaseURL + path
	if len(opts.Query) > 0 {
		fullURL += "?" + opts.Query.Encode()
	}

	var lastErr error
	for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("sqlacc: %s %s aborted: %w", method, fullURL, err)
		}

		attemptCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
		req, err := http.NewRequestWithContext(attemptCtx, method, fullURL, nil)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("sqlacc: build %s %s: %w", method, fullURL, err)
		}
		if len(body) > 0 {
			req.Body = io.NopCloser(bytes.NewReader(body))
			req.ContentLength = int64(len(body))
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("Accept", "application/json")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			cancel()
			lastErr = fmt.Errorf("sqlacc: %s %s attempt %d/%d: %w", method, fullURL, attempt, opts.MaxAttempts, err)
			if !shouldRetryErr(err) || attempt == opts.MaxAttempts {
				return nil, lastErr
			}
			sleepBackoff(attempt)
			continue
		}

		if shouldRetryStatus(resp.StatusCode) && attempt < opts.MaxAttempts {
			drainAndClose(resp)
			cancel()
			lastErr = fmt.Errorf("sqlacc: %s %s attempt %d/%d returned %d", method, fullURL, attempt, opts.MaxAttempts, resp.StatusCode)
			sleepBackoff(attempt)
			continue
		}

		resp.Body = &cancelOnClose{ReadCloser: resp.Body, cancel: cancel}
		return resp, nil
	}
	return nil, lastErr
}

type cancelOnClose struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnClose) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

func shouldRetryErr(err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}
	return true
}

func shouldRetryStatus(code int) bool {
	return code == http.StatusTooManyRequests || (code >= 500 && code <= 599)
}

func sleepBackoff(attempt int) {
	d := backoffBase * time.Duration(1<<(attempt-1))
	if d > backoffCap {
		d = backoffCap
	}
	jitter := time.Duration(rand.Int63n(int64(d / 2)))
	time.Sleep(d + jitter)
}

func drainAndClose(resp *http.Response) {
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

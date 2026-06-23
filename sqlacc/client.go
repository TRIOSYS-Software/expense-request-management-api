package sqlacc

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"shwetaik-expense-management-api/configs"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

const (
	emptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	defaultReadTimeout  = 30 * time.Second
	defaultWriteTimeout = 60 * time.Second
	defaultMaxAttempts  = 4
	backoffBase         = 500 * time.Millisecond
	backoffCap          = 8 * time.Second
)

// SQLAccClient is the shared HTTP client for the official SQL Acc REST API.
// It handles SigV4 signing, per-attempt timeouts, retry with exponential
// backoff for transient failures, and connection reuse.
type SQLAccClient struct {
	BaseURL    string
	Region     string
	Service    string
	AccessKey  string
	SecretKey  string
	HTTPClient *http.Client
	signer     *v4.Signer
}

// Default is the process-wide singleton. Initialized lazily on first use so
// configs.Envs is fully loaded by the time we read it.
var (
	defaultOnce   sync.Once
	defaultClient *SQLAccClient
)

// Default returns the shared client, constructing it on first call.
func Default() *SQLAccClient {
	defaultOnce.Do(func() {
		defaultClient = New(
			configs.Envs.SQLACC_API_URL,
			configs.Envs.SQLACC_API_REGION,
			configs.Envs.SQLACC_API_SERVICE,
			configs.Envs.SQLACC_ACCESS_KEY,
			configs.Envs.SQLACC_SECRET_KEY,
		)
	})
	return defaultClient
}

// New builds a SQLAccClient. Most callers should use Default() instead.
func New(baseURL, region, service, accessKey, secretKey string) *SQLAccClient {
	return &SQLAccClient{
		BaseURL:   baseURL,
		Region:    region,
		Service:   service,
		AccessKey: accessKey,
		SecretKey: secretKey,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		},
		signer: v4.NewSigner(),
	}
}

// RequestOpts configures per-request behavior.
type RequestOpts struct {
	// Timeout per attempt (default 30s for reads, 60s for writes).
	Timeout time.Duration
	// MaxAttempts including the first try (default 4 for reads, 1 for writes).
	MaxAttempts int
	// RetryWrites enables retry on POST/PUT/DELETE. Off by default to avoid
	// duplicate side effects; rely on idempotency at the caller instead.
	RetryWrites bool
	// Query is appended to the path; values are escaped.
	Query url.Values
}

// Get is a convenience wrapper for GET requests.
func (c *SQLAccClient) Get(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil, RequestOpts{Query: query})
}

// Post is a convenience wrapper for POST requests with a JSON body.
func (c *SQLAccClient) Post(ctx context.Context, path string, body []byte) (*http.Response, error) {
	return c.Do(ctx, http.MethodPost, path, body, RequestOpts{})
}

// Do sends an HTTP request to SQL Acc with SigV4 signing and resilience.
// body must be the full payload bytes (not a stream) so the request can be
// rebuilt on each retry attempt. Caller must close resp.Body.
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

	payloadHash := emptyPayloadHash
	if len(body) > 0 {
		sum := sha256.Sum256(body)
		payloadHash = hex.EncodeToString(sum[:])
	}

	creds := aws.Credentials{
		AccessKeyID:     c.AccessKey,
		SecretAccessKey: c.SecretKey,
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

		if err := c.signer.SignHTTP(attemptCtx, creds, req, payloadHash, c.Service, c.Region, time.Now().UTC()); err != nil {
			cancel()
			return nil, fmt.Errorf("sqlacc: sign %s %s: %w", method, fullURL, err)
		}

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

		// Wire the per-attempt cancel into the response body so it fires when
		// the caller closes the body (either via defer or on early return).
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

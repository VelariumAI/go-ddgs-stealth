package goddgs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Fetcher is a unified fetch interface for both HTTP and browser-backed implementations.
type Fetcher interface {
	Fetch(ctx context.Context, req FetchRequest) (*FetchResponse, error)
}

// StealthFetcher extends Fetcher with lifecycle control for browser-backed fetchers.
type StealthFetcher interface {
	Fetcher
	Close() error
}

// FetchRequest represents a single web fetch operation.
type FetchRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

// FetchResponse is a normalized result returned by all Fetcher implementations.
type FetchResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	FinalURL   string
	Duration   time.Duration
	Fetcher    string
}

// HTTPFetcher is a transport-backed implementation that reuses existing anti-bot primitives.
type HTTPFetcher struct {
	opts    StealthOptions
	client  *http.Client
	antiBot *antiBotState
}

var _ Fetcher = (*HTTPFetcher)(nil)

// NewHTTPFetcher builds a Fetcher that inherits go-ddgs anti-bot features when configured.
func NewHTTPFetcher(opts StealthOptions) (*HTTPFetcher, error) {
	opts = opts.withDefaults()

	var (
		st  *antiBotState
		hc  *http.Client
		err error
	)
	if opts.AntiBotConfig != nil {
		st, hc, err = buildAntiBotState(opts.AntiBotConfig)
		if err != nil {
			return nil, fmt.Errorf("build antibot state: %w", err)
		}
	}

	if opts.HTTPClient != nil {
		hc = opts.HTTPClient
	}
	if hc == nil {
		hc = &http.Client{Timeout: opts.RequestTimeout}
	}
	if opts.RequestTimeout > 0 {
		hc.Timeout = opts.RequestTimeout
	}

	return &HTTPFetcher{opts: opts, client: hc, antiBot: st}, nil
}

// Fetch executes an HTTP request with shared anti-bot protections.
func (f *HTTPFetcher) Fetch(ctx context.Context, req FetchRequest) (*FetchResponse, error) {
	return f.fetch(ctx, req, true)
}

func (f *HTTPFetcher) fetch(ctx context.Context, req FetchRequest, solverRetry bool) (*FetchResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	rawURL := strings.TrimSpace(req.URL)
	if rawURL == "" {
		return nil, fmt.Errorf("fetch url is required")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid fetch url: %q", rawURL)
	}

	if f.antiBot != nil && f.antiBot.circuit != nil && f.antiBot.circuit.IsOpen() {
		return nil, ErrCircuitOpen
	}
	if f.antiBot != nil && f.antiBot.rateLimit != nil {
		if err := f.antiBot.rateLimit.Wait(ctx); err != nil {
			return nil, err
		}
	}
	if f.antiBot != nil && f.antiBot.session != nil && f.antiBot.cfg != nil && f.antiBot.cfg.SessionWarmup && f.antiBot.session.NeedsWarmup() {
		home := parsed.Scheme + "://" + parsed.Host + "/"
		_ = f.antiBot.session.Warmup(ctx, f.client, home, func(ctx context.Context, method, warmURL string) (*http.Request, error) {
			r, err := http.NewRequestWithContext(ctx, method, warmURL, nil)
			if err != nil {
				return nil, err
			}
			for k, v := range f.opts.ExtraHeaders {
				r.Header.Set(k, v)
			}
			if f.opts.UserAgent != "" {
				r.Header.Set("User-Agent", f.opts.UserAgent)
			}
			return r, nil
		})
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, rawURL, bytes.NewReader(req.Body))
	if err != nil {
		return nil, err
	}
	for k, v := range f.opts.ExtraHeaders {
		httpReq.Header.Set(k, v)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}
	if f.opts.UserAgent != "" && httpReq.Header.Get("User-Agent") == "" {
		httpReq.Header.Set("User-Agent", f.opts.UserAgent)
	}

	started := time.Now()
	httpResp, err := f.client.Do(httpReq)
	if err != nil {
		if f.antiBot != nil && f.antiBot.rateLimit != nil {
			f.antiBot.rateLimit.OnBlock()
		}
		return nil, err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(httpResp.Body, f.opts.MaxBodyBytes))
	if err != nil {
		return nil, err
	}

	info := DetectBlockSignal(httpResp.StatusCode, httpResp.Header, body)
	if info.IsDetected() {
		if f.antiBot != nil && f.antiBot.rateLimit != nil {
			f.antiBot.rateLimit.OnBlock()
		}
		if f.antiBot != nil && f.antiBot.circuit != nil {
			f.antiBot.circuit.RecordBlock()
		}
		evt := BlockedEvent{StatusCode: httpResp.StatusCode, Headers: httpResp.Header.Clone(), BodySnippet: string(bodySnippet(body, 512)), Detector: info.DetectorKey, Attempt: 1}
		if f.antiBot != nil && f.antiBot.cfg != nil && f.antiBot.cfg.SessionInvalidateOnBlock && f.antiBot.session != nil {
			f.antiBot.session.Invalidate()
		}
		if solverRetry && f.antiBot != nil && f.antiBot.solver != nil {
			if _, solvedErr := f.antiBot.solver.Solve(ctx, rawURL, info, body); solvedErr == nil {
				return f.fetch(ctx, req, false)
			}
		}
		return nil, &BlockedError{Event: evt}
	}

	if f.antiBot != nil && f.antiBot.rateLimit != nil {
		f.antiBot.rateLimit.OnSuccess()
	}
	if f.antiBot != nil && f.antiBot.circuit != nil {
		f.antiBot.circuit.RecordSuccess()
	}

	finalURL := rawURL
	if httpResp.Request != nil && httpResp.Request.URL != nil {
		finalURL = httpResp.Request.URL.String()
	}
	return &FetchResponse{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header.Clone(),
		Body:       body,
		FinalURL:   finalURL,
		Duration:   time.Since(started),
		Fetcher:    "http",
	}, nil
}

func bodySnippet(body []byte, max int) []byte {
	if max <= 0 || len(body) <= max {
		return body
	}
	return body[:max]
}

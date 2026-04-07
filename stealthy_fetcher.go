package goddgs

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

var launchStealthBrowser = func(opts StealthOptions) (*rod.Browser, string, error) {
	l := launcher.New().Headless(opts.Headless)
	if dir := strings.TrimSpace(opts.PersistentContextDir); dir != "" {
		l = l.UserDataDir(dir)
	}
	if bin := strings.TrimSpace(opts.BrowserBinary); bin != "" {
		l = l.Bin(bin)
	}
	proxy := strings.TrimSpace(opts.ProxyURL)
	if proxy == "" && opts.AntiBotConfig != nil && opts.AntiBotConfig.ProxyPool != nil {
		if pe := opts.AntiBotConfig.ProxyPool.Next(); pe != nil {
			proxy = pe.URL
		}
	}
	if proxy != "" {
		l = l.Proxy(proxy)
	}
	launchURL, err := l.Launch()
	if err != nil {
		return nil, "", err
	}
	b := rod.New().ControlURL(launchURL)
	if err := b.Connect(); err != nil {
		return nil, "", err
	}
	return b, launchURL, nil
}

var runStealthPage = func(browser *rod.Browser, opts StealthOptions, req FetchRequest) (string, string, error) {
	var (
		outBody  string
		finalURL string
	)
	pageError := rod.Try(func() {
		incognito := browser.MustIncognito()
		page := incognito.MustPage("about:blank")
		defer page.MustClose()
		defer incognito.MustClose()

		if opts.UserAgent != "" {
			page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{UserAgent: opts.UserAgent})
		}
		if len(opts.ExtraHeaders) > 0 || len(req.Headers) > 0 {
			pairs := make([]string, 0, (len(opts.ExtraHeaders)+len(req.Headers))*2)
			for k, v := range opts.ExtraHeaders {
				pairs = append(pairs, k, v)
			}
			for k, v := range req.Headers {
				pairs = append(pairs, k, v)
			}
			page.MustSetExtraHeaders(pairs...)
		}
		for _, script := range StealthScripts(opts.StealthLevel) {
			page.MustEvalOnNewDocument(script)
		}
		if opts.HumanLikeBehavior {
			time.Sleep(time.Duration(120+rand.Intn(260)) * time.Millisecond)
		}
		page.MustNavigate(req.URL)
		page.MustWaitLoad()
		if opts.HumanLikeBehavior {
			time.Sleep(time.Duration(60+rand.Intn(180)) * time.Millisecond)
		}
		outBody = page.MustHTML()
		finalURL = page.MustInfo().URL
	})
	if pageError != nil {
		return "", "", pageError
	}
	return outBody, finalURL, nil
}

var closeStealthBrowser = func(b *rod.Browser) error { return b.Close() }

// StealthyFetcher is a browser-backed Fetcher built on Rod.
// It shares anti-bot pacing/circuit/solver/session logic via StealthOptions.
type StealthyFetcher struct {
	opts StealthOptions

	fallback *HTTPFetcher
	antiBot  *antiBotState

	mu         sync.Mutex
	launchURL  string
	browser    *rod.Browser
	launchedAt time.Time
}

var _ Fetcher = (*StealthyFetcher)(nil)

func NewStealthyFetcher(opts StealthOptions) (*StealthyFetcher, error) {
	opts = opts.withDefaults()
	hf, err := NewHTTPFetcher(opts)
	if err != nil {
		return nil, err
	}
	return &StealthyFetcher{
		opts:     opts,
		fallback: hf,
		antiBot:  hf.antiBot,
	}, nil
}

func (f *StealthyFetcher) Fetch(ctx context.Context, req FetchRequest) (*FetchResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(req.URL) == "" {
		return nil, fmt.Errorf("fetch url is required")
	}
	if f.antiBot != nil && f.antiBot.circuit != nil && f.antiBot.circuit.IsOpen() {
		return nil, ErrCircuitOpen
	}
	if f.antiBot != nil && f.antiBot.rateLimit != nil {
		if err := f.antiBot.rateLimit.Wait(ctx); err != nil {
			return nil, err
		}
	}

	browser, err := f.ensureBrowser()
	if err != nil {
		return f.fallback.Fetch(ctx, req)
	}

	started := time.Now()
	outBody, finalURL, fetchErr := runStealthPage(browser, f.opts, req)
	if fetchErr != nil {
		if f.antiBot != nil && f.antiBot.rateLimit != nil {
			f.antiBot.rateLimit.OnBlock()
		}
		return f.fallback.Fetch(ctx, req)
	}

	body := []byte(outBody)
	info := DetectBlockSignal(http.StatusOK, http.Header{}, body)
	if info.IsDetected() {
		if f.antiBot != nil && f.antiBot.rateLimit != nil {
			f.antiBot.rateLimit.OnBlock()
		}
		if f.antiBot != nil && f.antiBot.circuit != nil {
			f.antiBot.circuit.RecordBlock()
		}
		evt := BlockedEvent{StatusCode: http.StatusOK, Headers: http.Header{}, BodySnippet: string(bodySnippet(body, 512)), Detector: info.DetectorKey, Attempt: 1}
		if f.antiBot != nil && f.antiBot.solver != nil {
			if _, err := f.antiBot.solver.Solve(ctx, req.URL, info, body); err == nil {
				return f.fallback.Fetch(ctx, req)
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
	if finalURL == "" {
		finalURL = req.URL
	}

	return &FetchResponse{
		StatusCode: http.StatusOK,
		Headers:    http.Header{},
		Body:       body,
		FinalURL:   finalURL,
		Duration:   time.Since(started),
		Fetcher:    "stealth",
	}, nil
}

func (f *StealthyFetcher) ensureBrowser() (*rod.Browser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.browser != nil {
		return f.browser, nil
	}

	b, launchURL, err := launchStealthBrowser(f.opts)
	if err != nil {
		return nil, err
	}
	f.browser = b
	f.launchURL = launchURL
	f.launchedAt = time.Now()
	return b, nil
}

func (f *StealthyFetcher) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.browser == nil {
		return nil
	}
	err := closeStealthBrowser(f.browser)
	f.browser = nil
	f.launchURL = ""
	f.launchedAt = time.Time{}
	return err
}

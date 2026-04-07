package goddgs

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/net/html"
)

type serviceStealthFetchRequest struct {
	URL      string            `json:"url"`
	Method   string            `json:"method,omitempty"`
	Mode     string            `json:"mode,omitempty"` // http|stealth
	Headers  map[string]string `json:"headers,omitempty"`
	Body     []byte            `json:"body,omitempty"`
	Human    bool              `json:"human_like,omitempty"`
	Level    string            `json:"stealth_level,omitempty"`
	TimeoutS int               `json:"timeout_seconds,omitempty"`
}

type serviceStealthCrawlRequest struct {
	StartURL     string `json:"start_url"`
	MaxPages     int    `json:"max_pages,omitempty"`
	Concurrency  int    `json:"concurrency,omitempty"`
	Checkpoint   string `json:"checkpoint_file,omitempty"`
	StreamingOut string `json:"streaming_jsonl,omitempty"`
	TimeoutS     int    `json:"timeout_seconds,omitempty"`
}

type serviceStealthCrawlResponse struct {
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	Status    string    `json:"status"`
}

func registerStealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/v1/stealth/fetch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !requireAPIToken(r) {
			writeAPIErr(w, http.StatusUnauthorized, APIError{Error: "unauthorized", Kind: string(ErrKindInvalidInput)})
			return
		}
		if !stealthGuard.allow(requesterIP(r)) {
			writeAPIErr(w, http.StatusTooManyRequests, APIError{Error: "rate limit exceeded", Kind: string(ErrKindRateLimited)})
			return
		}
		var req serviceStealthFetchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeAPIErr(w, http.StatusBadRequest, APIError{Error: "invalid json", Kind: string(ErrKindInvalidInput)})
			return
		}
		if strings.TrimSpace(req.URL) == "" {
			writeAPIErr(w, http.StatusBadRequest, APIError{Error: "url is required", Kind: string(ErrKindInvalidInput)})
			return
		}
		timeout := 30 * time.Second
		if req.TimeoutS > 0 {
			timeout = time.Duration(req.TimeoutS) * time.Second
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		ctx, span := startSpan(ctx, "http.stealth.fetch", attribute.String("mode", req.Mode), attribute.String("url", req.URL))
		defer endSpan(span, nil)

		opts := StealthOptions{
			AntiBotConfig:     NewAntiBotConfig(),
			HumanLikeBehavior: req.Human,
			StealthLevel:      StealthLevel(strings.ToLower(strings.TrimSpace(req.Level))),
		}
		if opts.StealthLevel == "" {
			opts.StealthLevel = StealthLevelStrong
		}

		var fetcher Fetcher
		if strings.EqualFold(strings.TrimSpace(req.Mode), "stealth") {
			sf, err := NewStealthyFetcher(opts)
			if err != nil {
				writeAPIErr(w, http.StatusBadGateway, APIError{Error: "stealth fetcher init failed: " + err.Error(), Kind: string(ErrKindProviderUnavailable)})
				return
			}
			defer sf.Close()
			fetcher = sf
		} else {
			hf, err := NewHTTPFetcher(opts)
			if err != nil {
				writeAPIErr(w, http.StatusBadGateway, APIError{Error: "http fetcher init failed: " + err.Error(), Kind: string(ErrKindProviderUnavailable)})
				return
			}
			fetcher = hf
		}

		resp, err := fetcher.Fetch(ctx, FetchRequest{Method: req.Method, URL: req.URL, Headers: req.Headers, Body: req.Body})
		if err != nil {
			if IsBlocked(err) {
				writeAPIErr(w, http.StatusTooManyRequests, APIError{Error: err.Error(), Kind: string(ErrKindBlocked)})
				return
			}
			writeAPIErr(w, http.StatusBadGateway, APIError{Error: err.Error(), Kind: string(ErrKindProviderUnavailable)})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/v1/stealth/crawl", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !requireAPIToken(r) {
			writeAPIErr(w, http.StatusUnauthorized, APIError{Error: "unauthorized", Kind: string(ErrKindInvalidInput)})
			return
		}
		if !stealthGuard.allow(requesterIP(r)) {
			writeAPIErr(w, http.StatusTooManyRequests, APIError{Error: "rate limit exceeded", Kind: string(ErrKindRateLimited)})
			return
		}
		var req serviceStealthCrawlRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeAPIErr(w, http.StatusBadRequest, APIError{Error: "invalid json", Kind: string(ErrKindInvalidInput)})
			return
		}
		if strings.TrimSpace(req.StartURL) == "" {
			writeAPIErr(w, http.StatusBadRequest, APIError{Error: "start_url is required", Kind: string(ErrKindInvalidInput)})
			return
		}

		hf, err := NewHTTPFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig()})
		if err != nil {
			writeAPIErr(w, http.StatusBadGateway, APIError{Error: err.Error(), Kind: string(ErrKindProviderUnavailable)})
			return
		}

		started := time.Now().UTC()
		spider, err := NewSpider(hf, SpiderConfig{
			StartURLs:          []string{req.StartURL},
			Concurrency:        req.Concurrency,
			MaxPages:           req.MaxPages,
			CheckpointFileJSON: strings.TrimSpace(req.Checkpoint),
			StreamingJSONL:     strings.TrimSpace(req.StreamingOut),
			Parse: func(_ context.Context, _ CrawlResult, body []byte) ([]string, error) {
				return parseAnchors(body)
			},
		})
		if err != nil {
			writeAPIErr(w, http.StatusBadGateway, APIError{Error: err.Error(), Kind: string(ErrKindProviderUnavailable)})
			return
		}
		defer spider.Close()

		crawlTimeout := 2 * time.Minute
		if req.TimeoutS > 0 {
			crawlTimeout = time.Duration(req.TimeoutS) * time.Second
		}
		ctx, cancel := context.WithTimeout(r.Context(), crawlTimeout)
		defer cancel()
		ctx, span := startSpan(ctx, "http.stealth.crawl", attribute.String("start_url", req.StartURL))
		err = spider.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			endSpan(span, err)
			writeAPIErr(w, http.StatusBadGateway, APIError{Error: err.Error(), Kind: string(ErrKindProviderUnavailable)})
			return
		}
		endSpan(span, nil)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(serviceStealthCrawlResponse{StartedAt: started, EndedAt: time.Now().UTC(), Status: "ok"})
	})
}

func parseAnchors(body []byte) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	links := make([]string, 0, 32)
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, "a") {
			for _, a := range n.Attr {
				if strings.EqualFold(a.Key, "href") {
					links = append(links, a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			visit(c)
		}
	}
	visit(doc)
	return links, nil
}

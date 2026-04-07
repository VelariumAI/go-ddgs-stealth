package goddgs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// CrawlResult is emitted per crawled URL.
type CrawlResult struct {
	URL        string            `json:"url"`
	FinalURL   string            `json:"final_url,omitempty"`
	StatusCode int               `json:"status_code"`
	Discovered []string          `json:"discovered,omitempty"`
	Fetcher    string            `json:"fetcher,omitempty"`
	Error      string            `json:"error,omitempty"`
	Meta       map[string]string `json:"meta,omitempty"`
	At         time.Time         `json:"at"`
}

// ParseFunc inspects a fetched page and returns discovered URLs.
type ParseFunc func(ctx context.Context, result CrawlResult, body []byte) ([]string, error)

// SpiderConfig controls crawling behavior.
type SpiderConfig struct {
	StartURLs          []string
	Concurrency        int
	DomainMinInterval  time.Duration
	CheckpointFileJSON string
	StreamingJSONL     string
	MaxPages           int
	MaxDepth           int
	AllowHosts         []string
	DenyHosts          []string
	Parse              ParseFunc
}

// Spider is a concurrency-safe crawler with checkpoint/resume support.
type Spider struct {
	fetcher Fetcher
	cfg     SpiderConfig

	domainMu  sync.Mutex
	domainHit map[string]time.Time

	seenMu sync.Mutex
	seen   map[string]struct{}

	writerMu sync.Mutex
	writer   *jsonlWriter

	pauseMu sync.RWMutex
	paused  bool

	allowHosts map[string]struct{}
	denyHosts  map[string]struct{}
}

func NewSpider(fetcher Fetcher, cfg SpiderConfig) (*Spider, error) {
	if fetcher == nil {
		return nil, fmt.Errorf("fetcher is required")
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 4
	}
	if cfg.DomainMinInterval <= 0 {
		cfg.DomainMinInterval = 350 * time.Millisecond
	}
	if cfg.MaxPages <= 0 {
		cfg.MaxPages = 200
	}
	if cfg.MaxDepth <= 0 {
		cfg.MaxDepth = 4
	}
	if len(cfg.StartURLs) == 0 {
		return nil, fmt.Errorf("at least one start url is required")
	}
	sp := &Spider{
		fetcher:    fetcher,
		cfg:        cfg,
		domainHit:  map[string]time.Time{},
		seen:       map[string]struct{}{},
		allowHosts: normalizeHostSet(cfg.AllowHosts),
		denyHosts:  normalizeHostSet(cfg.DenyHosts),
	}
	if strings.TrimSpace(cfg.StreamingJSONL) != "" {
		w, err := newJSONLWriter(cfg.StreamingJSONL)
		if err != nil {
			return nil, err
		}
		sp.writer = w
	}
	_ = sp.restoreCheckpoint()
	return sp, nil
}

func (s *Spider) Close() error {
	if s.writer != nil {
		return s.writer.Close()
	}
	return nil
}

func (s *Spider) Run(ctx context.Context) error {
	type crawlTask struct {
		url   string
		depth int
	}
	queue := make(chan crawlTask, s.cfg.Concurrency*8)
	out := make(chan CrawlResult, s.cfg.Concurrency*4)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pageBudget := s.cfg.MaxPages
	var budgetMu sync.Mutex
	consumeBudget := func() bool {
		budgetMu.Lock()
		defer budgetMu.Unlock()
		if pageBudget <= 0 {
			return false
		}
		pageBudget--
		return true
	}

	for i := 0; i < s.cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-queue:
					if !ok {
						return
					}
					if err := s.waitIfPaused(ctx); err != nil {
						return
					}
					if !consumeBudget() {
						cancel()
						return
					}
					res := s.fetchOne(ctx, task.url)
					if res.Meta == nil {
						res.Meta = map[string]string{}
					}
					res.Meta["depth"] = fmt.Sprintf("%d", task.depth)
					select {
					case out <- res:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	seeded := 0
	for _, u := range s.cfg.StartURLs {
		if !s.urlAllowed(u) {
			continue
		}
		if s.markSeen(u) {
			queue <- crawlTask{url: u, depth: 0}
			seeded++
		}
	}
	if seeded == 0 {
		close(queue)
		wg.Wait()
		close(out)
		return nil
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	inFlight := seeded
	for res := range out {
		inFlight--
		s.persistResult(res)
		currentDepth := 0
		if res.Meta != nil {
			if d, ok := res.Meta["depth"]; ok {
				_, _ = fmt.Sscanf(d, "%d", &currentDepth)
			}
		}
		nextDepth := currentDepth + 1

		for _, next := range res.Discovered {
			if nextDepth > s.cfg.MaxDepth {
				continue
			}
			if !s.urlAllowed(next) {
				continue
			}
			if !s.markSeen(next) {
				continue
			}
			select {
			case queue <- crawlTask{url: next, depth: nextDepth}:
				inFlight++
			case <-ctx.Done():
			}
		}
		if inFlight <= 0 {
			break
		}
	}
	close(queue)
	return ctx.Err()
}

func normalizeHostSet(hosts []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, h := range hosts {
		h = strings.ToLower(strings.TrimSpace(h))
		if h != "" {
			out[h] = struct{}{}
		}
	}
	return out
}

func (s *Spider) urlAllowed(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Host)
	if len(s.allowHosts) > 0 {
		if _, ok := s.allowHosts[host]; !ok {
			return false
		}
	}
	if _, denied := s.denyHosts[host]; denied {
		return false
	}
	return true
}

func (s *Spider) fetchOne(ctx context.Context, rawURL string) CrawlResult {
	s.waitDomain(rawURL)
	res := CrawlResult{URL: rawURL, At: time.Now().UTC()}

	fr, err := s.fetcher.Fetch(ctx, FetchRequest{Method: "GET", URL: rawURL})
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.StatusCode = fr.StatusCode
	res.FinalURL = fr.FinalURL
	res.Fetcher = fr.Fetcher

	if s.cfg.Parse != nil {
		discovered, err := s.cfg.Parse(ctx, res, fr.Body)
		if err != nil {
			res.Error = err.Error()
		} else {
			res.Discovered = normalizeURLs(fr.FinalURL, discovered)
		}
	}
	return res
}

// Pause stops worker progress until Resume is called.
func (s *Spider) Pause() {
	s.pauseMu.Lock()
	s.paused = true
	s.pauseMu.Unlock()
}

// Resume re-enables worker progress.
func (s *Spider) Resume() {
	s.pauseMu.Lock()
	s.paused = false
	s.pauseMu.Unlock()
}

// IsPaused reports whether the spider is currently paused.
func (s *Spider) IsPaused() bool {
	s.pauseMu.RLock()
	defer s.pauseMu.RUnlock()
	return s.paused
}

func (s *Spider) waitIfPaused(ctx context.Context) error {
	for {
		if !s.IsPaused() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func (s *Spider) waitDomain(rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return
	}
	host := strings.ToLower(u.Host)
	s.domainMu.Lock()
	defer s.domainMu.Unlock()
	if ts, ok := s.domainHit[host]; ok {
		if wait := s.cfg.DomainMinInterval - time.Since(ts); wait > 0 {
			time.Sleep(wait)
		}
	}
	s.domainHit[host] = time.Now()
}

func (s *Spider) markSeen(rawURL string) bool {
	key := strings.TrimSpace(rawURL)
	if key == "" {
		return false
	}
	s.seenMu.Lock()
	defer s.seenMu.Unlock()
	if _, exists := s.seen[key]; exists {
		return false
	}
	s.seen[key] = struct{}{}
	return true
}

func normalizeURLs(base string, links []string) []string {
	out := make([]string, 0, len(links))
	baseURL, _ := url.Parse(base)
	for _, link := range links {
		link = strings.TrimSpace(link)
		if link == "" {
			continue
		}
		u, err := url.Parse(link)
		if err != nil {
			continue
		}
		if !u.IsAbs() && baseURL != nil {
			u = baseURL.ResolveReference(u)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			continue
		}
		out = append(out, u.String())
	}
	return out
}

func (s *Spider) persistResult(res CrawlResult) {
	if s.writer != nil {
		s.writerMu.Lock()
		_ = s.writer.Write(res)
		s.writerMu.Unlock()
	}
	if strings.TrimSpace(s.cfg.CheckpointFileJSON) == "" {
		return
	}
	tmp := s.cfg.CheckpointFileJSON + ".tmp"
	checkpoint := struct {
		Seen []string `json:"seen"`
	}{Seen: s.snapshotSeen()}
	f, err := os.Create(tmp)
	if err != nil {
		return
	}
	_ = json.NewEncoder(f).Encode(checkpoint)
	_ = f.Close()
	_ = os.Rename(tmp, s.cfg.CheckpointFileJSON)
}

func (s *Spider) snapshotSeen() []string {
	s.seenMu.Lock()
	defer s.seenMu.Unlock()
	out := make([]string, 0, len(s.seen))
	for u := range s.seen {
		out = append(out, u)
	}
	return out
}

func (s *Spider) restoreCheckpoint() error {
	path := strings.TrimSpace(s.cfg.CheckpointFileJSON)
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	var checkpoint struct {
		Seen []string `json:"seen"`
	}
	if err := json.NewDecoder(f).Decode(&checkpoint); err != nil {
		return err
	}
	for _, u := range checkpoint.Seen {
		s.seen[strings.TrimSpace(u)] = struct{}{}
	}
	return nil
}

type jsonlWriter struct {
	mu sync.Mutex
	f  *os.File
	w  *bufio.Writer
}

func newJSONLWriter(path string) (*jsonlWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &jsonlWriter{f: f, w: bufio.NewWriterSize(f, 64*1024)}, nil
}

func (j *jsonlWriter) Write(v any) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := j.w.Write(append(b, '\n')); err != nil {
		return err
	}
	return j.w.Flush()
}

func (j *jsonlWriter) Close() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.w != nil {
		_ = j.w.Flush()
	}
	if j.f != nil {
		return j.f.Close()
	}
	return nil
}

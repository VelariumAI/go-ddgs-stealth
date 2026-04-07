package goddgs

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServiceStealthFetchEndpoint(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer target.Close()

	p := &fakeProvider{name: "ddg", enabled: true, fn: func(_ context.Context, _ SearchRequest) ([]Result, error) {
		return []Result{{Title: "x", URL: "https://x"}}, nil
	}}
	engine, _ := NewEngine(EngineOptions{Providers: []Provider{p}})
	h := NewHTTPHandler(engine, Config{Timeout: 2 * time.Second}, nil)
	s := httptest.NewServer(h)
	defer s.Close()

	b, _ := json.Marshal(map[string]any{"url": target.URL, "mode": "http"})
	resp, err := http.Post(s.URL+"/v1/stealth/fetch", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestServiceStealthCrawlEndpoint(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<a href="/next">n</a>`))
	}))
	defer target.Close()

	p := &fakeProvider{name: "ddg", enabled: true, fn: func(_ context.Context, _ SearchRequest) ([]Result, error) {
		return []Result{{Title: "x", URL: "https://x"}}, nil
	}}
	engine, _ := NewEngine(EngineOptions{Providers: []Provider{p}})
	h := NewHTTPHandler(engine, Config{Timeout: 2 * time.Second}, nil)
	s := httptest.NewServer(h)
	defer s.Close()

	b, _ := json.Marshal(map[string]any{"start_url": target.URL, "max_pages": 2, "concurrency": 1})
	resp, err := http.Post(s.URL+"/v1/stealth/crawl", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

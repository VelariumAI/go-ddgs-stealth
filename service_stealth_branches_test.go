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

func makeServiceServer(t *testing.T) *httptest.Server {
	t.Helper()
	p := &fakeProvider{name: "ddg", enabled: true, fn: func(_ context.Context, _ SearchRequest) ([]Result, error) {
		return []Result{{Title: "x", URL: "https://x"}}, nil
	}}
	engine, _ := NewEngine(EngineOptions{Providers: []Provider{p}})
	return httptest.NewServer(NewHTTPHandler(engine, Config{Timeout: 2 * time.Second}, nil))
}

func TestStealthEndpointsMethodNotAllowed(t *testing.T) {
	s := makeServiceServer(t)
	defer s.Close()
	resp, _ := http.Get(s.URL + "/v1/stealth/fetch")
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("fetch status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()
	resp, _ = http.Get(s.URL + "/v1/stealth/crawl")
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("crawl status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestStealthFetchBlockedAndBadGatewayPaths(t *testing.T) {
	blocked := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("access denied"))
	}))
	defer blocked.Close()

	s := makeServiceServer(t)
	defer s.Close()

	b, _ := json.Marshal(map[string]any{"url": blocked.URL, "mode": "http", "timeout_seconds": 1})
	resp, err := http.Post(s.URL+"/v1/stealth/fetch", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("blocked status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	b, _ = json.Marshal(map[string]any{"url": "http://127.0.0.1:1", "mode": "http"})
	resp, err = http.Post(s.URL+"/v1/stealth/fetch", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("badgateway status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestStealthFetchStealthModeAndCrawlErrorBranch(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer target.Close()

	s := makeServiceServer(t)
	defer s.Close()

	b, _ := json.Marshal(map[string]any{"url": target.URL, "mode": "stealth"})
	resp, err := http.Post(s.URL+"/v1/stealth/fetch", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("stealth status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	b, _ = json.Marshal(map[string]any{"start_url": "http://127.0.0.1:1", "max_pages": 1})
	resp, err = http.Post(s.URL+"/v1/stealth/crawl", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("crawl status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestStealthCrawlTimeoutBranch(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1500 * time.Millisecond)
		_, _ = w.Write([]byte(`<a href="/next">n</a>`))
	}))
	defer slow.Close()

	s := makeServiceServer(t)
	defer s.Close()
	b, _ := json.Marshal(map[string]any{"start_url": slow.URL, "max_pages": 2, "timeout_seconds": 1})
	resp, err := http.Post(s.URL+"/v1/stealth/crawl", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("timeout status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

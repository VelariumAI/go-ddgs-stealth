package goddgs

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServiceStealthValidationPaths(t *testing.T) {
	p := &fakeProvider{name: "ddg", enabled: true, fn: func(_ context.Context, _ SearchRequest) ([]Result, error) {
		return []Result{{Title: "x", URL: "https://x"}}, nil
	}}
	engine, _ := NewEngine(EngineOptions{Providers: []Provider{p}})
	h := NewHTTPHandler(engine, Config{Timeout: 2 * time.Second}, nil)
	s := httptest.NewServer(h)
	defer s.Close()

	resp, err := http.Post(s.URL+"/v1/stealth/fetch", "application/json", bytes.NewBufferString("{"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("fetch invalid json status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	resp, err = http.Post(s.URL+"/v1/stealth/crawl", "application/json", bytes.NewBufferString(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("crawl missing start status=%d", resp.StatusCode)
	}
	_ = resp.Body.Close()
}

func TestParseAnchors(t *testing.T) {
	links, err := parseAnchors([]byte(`<a href="https://a">a</a><a href="/b">b</a>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 {
		t.Fatalf("links=%d want 2", len(links))
	}
}

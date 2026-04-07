package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	goddgs "github.com/velariumai/go-ddgs-stealth"
)

func TestRunStealthFetchHTTPMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
	defer srv.Close()

	code := runStealthFetch([]string{"--url", srv.URL, "--mode", "http", "--json"})
	if code != 0 {
		t.Fatalf("runStealthFetch code=%d want 0", code)
	}
}

func TestRunStealthCrawl(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			_, _ = w.Write([]byte(`<a href="/a">A</a><a href="/b">B</a>`))
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	out := filepath.Join(t.TempDir(), "crawl.jsonl")
	code := runStealthCrawl([]string{"--url", srv.URL, "--max", "3", "--concurrency", "2", "--out", out})
	if code != 0 {
		t.Fatalf("runStealthCrawl code=%d want 0", code)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		t.Fatal("expected non-empty crawl output")
	}
}

func TestParseHTMLAnchors(t *testing.T) {
	links, err := parseHTMLAnchors(context.Background(), goddgs.CrawlResult{}, []byte(`<a href="https://a">a</a><a href="/b">b</a>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 {
		t.Fatalf("links=%d want 2", len(links))
	}
}

func TestRunREPLQuit(t *testing.T) {
	old := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() { os.Stdin = old }()
	_, _ = w.WriteString("quit\n")
	_ = w.Close()

	if code := runREPL(); code != 0 {
		t.Fatalf("runREPL code=%d want 0", code)
	}
}

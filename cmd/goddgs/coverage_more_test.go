package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRunStealthFetchValidationAndFailure(t *testing.T) {
	if code := runStealthFetch([]string{}); code != 2 {
		t.Fatalf("missing url code=%d want 2", code)
	}
	if code := runStealthFetch([]string{"--url", "http://127.0.0.1:1", "--mode", "http"}); code != 3 {
		t.Fatalf("unreachable code=%d want 3", code)
	}
}

func TestRunStealthFetchStealthMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()
	if code := runStealthFetch([]string{"--url", srv.URL, "--mode", "stealth"}); code != 0 {
		t.Fatalf("stealth mode code=%d want 0", code)
	}
}

func TestRunStealthCrawlValidation(t *testing.T) {
	if code := runStealthCrawl([]string{}); code != 2 {
		t.Fatalf("missing url code=%d want 2", code)
	}
}

func TestRunREPLUnknownPath(t *testing.T) {
	old := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() { os.Stdin = old }()
	_, _ = w.WriteString("unknown command\nquit\n")
	_ = w.Close()
	if code := runREPL(); code != 0 {
		t.Fatalf("runREPL code=%d want 0", code)
	}
}

func TestRunREPLFetchPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	old := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() { os.Stdin = old }()
	_, _ = w.WriteString("fetch " + srv.URL + "\nquit\n")
	_ = w.Close()
	if code := runREPL(); code != 0 {
		t.Fatalf("runREPL code=%d want 0", code)
	}
}

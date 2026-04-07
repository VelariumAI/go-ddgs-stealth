package goddgs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPFetcherFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	f, err := NewHTTPFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig()})
	if err != nil {
		t.Fatal(err)
	}
	res, err := f.Fetch(context.Background(), FetchRequest{Method: "GET", URL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 || strings.TrimSpace(string(res.Body)) != "ok" {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestHTTPFetcherBlockedDetection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("access denied"))
	}))
	defer srv.Close()

	f, err := NewHTTPFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Fetch(context.Background(), FetchRequest{Method: "GET", URL: srv.URL})
	if err == nil || !IsBlocked(err) {
		t.Fatalf("expected blocked error, got %v", err)
	}
}

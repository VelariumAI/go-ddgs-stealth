package goddgs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewStealthyFetcherAndClose(t *testing.T) {
	sf, err := NewStealthyFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig(), BrowserBinary: "/no/such/browser"})
	if err != nil {
		t.Fatal(err)
	}
	if err := sf.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestStealthyFetcherFallbackWhenBrowserUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok from http fallback"))
	}))
	defer srv.Close()

	sf, err := NewStealthyFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig(), BrowserBinary: "/no/such/browser"})
	if err != nil {
		t.Fatal(err)
	}
	defer sf.Close()

	res, err := sf.Fetch(context.Background(), FetchRequest{Method: "GET", URL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if res.Fetcher != "http" {
		t.Fatalf("expected http fallback, got %q", res.Fetcher)
	}
}

package goddgs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-rod/rod"
)

func TestStealthyFetcherFetchWithPreseededBrowserFallsBack(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	sf, err := NewStealthyFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig(), BrowserBinary: "/no/such/browser"})
	if err != nil {
		t.Fatal(err)
	}

	// Pre-seed with a non-connected browser to execute the browser-attempt branch.
	sf.browser = &rod.Browser{}
	res, err := sf.Fetch(context.Background(), FetchRequest{Method: "GET", URL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if res.Fetcher != "http" {
		t.Fatalf("expected fallback fetcher=http, got %q", res.Fetcher)
	}
	sf.browser = nil
}

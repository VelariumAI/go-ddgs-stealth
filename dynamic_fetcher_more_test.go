package goddgs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDynamicFetcherFetchAndValidation(t *testing.T) {
	df, err := NewDynamicFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig(), BrowserBinary: "/no/such/browser"})
	if err != nil {
		t.Fatal(err)
	}
	defer df.Close()

	if _, err := df.Fetch(context.Background(), FetchRequest{URL: ""}); err == nil {
		t.Fatal("expected URL validation error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := df.FetchDynamic(ctx, DynamicFetchRequest{FetchRequest: FetchRequest{URL: "https://example.com"}, NetworkIdleWait: 100 * time.Millisecond}); err == nil {
		t.Fatal("expected context cancellation")
	}
}

func TestDynamicFetcherActions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><body><button id="go">go</button></body></html>`))
	}))
	defer srv.Close()

	df, err := NewDynamicFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig(), BrowserBinary: "/no/such/browser"})
	if err != nil {
		t.Fatal(err)
	}
	defer df.Close()

	_, err = df.FetchDynamic(context.Background(), DynamicFetchRequest{
		FetchRequest:    FetchRequest{URL: srv.URL},
		WaitForSelector: `id="go"`,
		Actions:         []DynamicAction{{Type: "wait", Selector: `id="go"`}, {Type: "click", Selector: `id="go"`}, {Type: "eval", Script: "1+1"}},
		NetworkIdleWait: 1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = df.FetchDynamic(context.Background(), DynamicFetchRequest{FetchRequest: FetchRequest{URL: srv.URL}, Actions: []DynamicAction{{Type: "unknown"}}})
	if err == nil {
		t.Fatal("expected unsupported action error")
	}

	_, err = df.FetchDynamic(context.Background(), DynamicFetchRequest{
		FetchRequest: FetchRequest{URL: srv.URL},
		Actions:      []DynamicAction{{Type: "wait"}},
	})
	if err == nil {
		t.Fatal("expected wait selector validation error")
	}

	_, err = df.FetchDynamic(context.Background(), DynamicFetchRequest{
		FetchRequest: FetchRequest{URL: srv.URL},
		Actions:      []DynamicAction{{Type: "eval"}},
	})
	if err == nil {
		t.Fatal("expected eval script validation error")
	}
}

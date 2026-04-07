package goddgs

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-rod/rod"
)

func TestStealthyFetcherFetchMockedSuccessAndBlock(t *testing.T) {
	origLaunch := launchStealthBrowser
	origRun := runStealthPage
	defer func() {
		launchStealthBrowser = origLaunch
		runStealthPage = origRun
	}()

	launchStealthBrowser = func(opts StealthOptions) (*rod.Browser, string, error) {
		return &rod.Browser{}, "ws://mock", nil
	}

	sf, err := NewStealthyFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig()})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { sf.browser = nil }()

	runStealthPage = func(_ *rod.Browser, _ StealthOptions, _ FetchRequest) (string, string, error) {
		return "<html><body>ok</body></html>", "https://example.com", nil
	}
	res, err := sf.Fetch(context.Background(), FetchRequest{URL: "https://example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Fetcher != "stealth" || res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected response: %+v", res)
	}

	runStealthPage = func(_ *rod.Browser, _ StealthOptions, _ FetchRequest) (string, string, error) {
		return "access denied", "https://example.com", nil
	}
	_, err = sf.Fetch(context.Background(), FetchRequest{URL: "https://example.com"})
	if err == nil || !IsBlocked(err) {
		t.Fatalf("expected blocked error, got %v", err)
	}
}

func TestStealthyFetcherFetchMockedFailureFallsBack(t *testing.T) {
	origLaunch := launchStealthBrowser
	origRun := runStealthPage
	defer func() {
		launchStealthBrowser = origLaunch
		runStealthPage = origRun
	}()

	launchStealthBrowser = func(opts StealthOptions) (*rod.Browser, string, error) {
		return nil, "", errors.New("boom")
	}

	sf, err := NewStealthyFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = sf.Fetch(context.Background(), FetchRequest{URL: "https://127.0.0.1:1"})
	if err == nil {
		t.Fatal("expected fallback error")
	}
}

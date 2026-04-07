package goddgs

import (
	"errors"
	"testing"

	"github.com/go-rod/rod"
)

func TestStealthyFetcherClosePaths(t *testing.T) {
	sf, err := NewStealthyFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig(), BrowserBinary: "/no/such/browser"})
	if err != nil {
		t.Fatal(err)
	}
	if err := sf.Close(); err != nil {
		t.Fatal(err)
	}

	orig := closeStealthBrowser
	defer func() { closeStealthBrowser = orig }()
	closeStealthBrowser = func(b *rod.Browser) error { return errors.New("close fail") }
	sf.browser = &rod.Browser{}
	if err := sf.Close(); err == nil {
		t.Fatal("expected close error")
	}
}

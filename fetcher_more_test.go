package goddgs

import (
	"context"
	"testing"
)

func TestHTTPFetcherInvalidURLAndBodySnippet(t *testing.T) {
	f, err := NewHTTPFetcher(StealthOptions{AntiBotConfig: NewAntiBotConfig()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Fetch(context.Background(), FetchRequest{URL: "not-a-url"}); err == nil {
		t.Fatal("expected invalid URL error")
	}
	if got := string(bodySnippet([]byte("abcdef"), 3)); got != "abc" {
		t.Fatalf("snippet=%q want abc", got)
	}
}

package goddgs

import (
	"context"
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

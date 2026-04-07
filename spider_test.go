package goddgs

import (
	"context"
	"testing"
)

func TestSpiderRunWithMockFetcher(t *testing.T) {
	f := mockFetcher("http")
	sp, err := NewSpider(f, SpiderConfig{
		StartURLs:   []string{"https://example.com"},
		Concurrency: 2,
		MaxPages:    3,
		Parse: func(_ context.Context, res CrawlResult, _ []byte) ([]string, error) {
			if res.URL == "https://example.com" {
				return []string{"https://example.com/a", "https://example.com/b"}, nil
			}
			return nil, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()

	err = sp.Run(context.Background())
	if err != nil && err != context.Canceled {
		t.Fatalf("run error: %v", err)
	}
	if got := len(sp.snapshotSeen()); got < 3 {
		t.Fatalf("seen=%d want >=3", got)
	}
}

func (m mockFetcher) String() string { return string(m) }

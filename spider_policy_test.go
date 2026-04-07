package goddgs

import (
	"context"
	"testing"
)

func TestSpiderPoliciesDepthAndHostFilters(t *testing.T) {
	sp, err := NewSpider(mockFetcher("http"), SpiderConfig{
		StartURLs:   []string{"https://example.com"},
		Concurrency: 1,
		MaxPages:    10,
		MaxDepth:    1,
		AllowHosts:  []string{"example.com"},
		DenyHosts:   []string{"denied.com"},
		Parse: func(_ context.Context, res CrawlResult, _ []byte) ([]string, error) {
			if res.URL == "https://example.com" {
				return []string{"https://example.com/a", "https://example.com/b", "https://denied.com/x"}, nil
			}
			return []string{"https://example.com/deeper"}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	_ = sp.Run(context.Background())

	seen := sp.snapshotSeen()
	foundDenied := false
	foundDeeper := false
	for _, u := range seen {
		if u == "https://denied.com/x" {
			foundDenied = true
		}
		if u == "https://example.com/deeper" {
			foundDeeper = true
		}
	}
	if foundDenied {
		t.Fatal("expected denied host to be filtered")
	}
	if foundDeeper {
		t.Fatal("expected deeper URL beyond max depth to be filtered")
	}
}

package goddgs

import (
	"context"
	"testing"
	"time"
)

func TestSpiderPauseResume(t *testing.T) {
	sp, err := NewSpider(mockFetcher("http"), SpiderConfig{
		StartURLs:   []string{"https://example.com"},
		Concurrency: 1,
		MaxPages:    2,
		Parse:       func(context.Context, CrawlResult, []byte) ([]string, error) { return nil, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()

	sp.Pause()
	if !sp.IsPaused() {
		t.Fatal("expected paused")
	}
	go func() {
		time.Sleep(80 * time.Millisecond)
		sp.Resume()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = sp.Run(ctx)
	if sp.IsPaused() {
		t.Fatal("expected resumed")
	}
}

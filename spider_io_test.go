package goddgs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJSONLWriterAndCheckpointRestore(t *testing.T) {
	tmp := t.TempDir()
	jsonl := filepath.Join(tmp, "crawl.jsonl")
	checkpoint := filepath.Join(tmp, "checkpoint.json")

	w, err := newJSONLWriter(jsonl)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Write(CrawlResult{URL: "https://example.com"}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(jsonl); err != nil || st.Size() == 0 {
		t.Fatalf("expected jsonl output, err=%v size=%d", err, st.Size())
	}

	sp, err := NewSpider(mockFetcher("http"), SpiderConfig{
		StartURLs:          []string{"https://example.com"},
		CheckpointFileJSON: checkpoint,
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = sp.Run(context.Background())
	sp.persistResult(CrawlResult{URL: "https://example.com", At: time.Now().UTC()})
	_ = sp.Close()

	sp2, err := NewSpider(mockFetcher("http"), SpiderConfig{
		StartURLs:          []string{"https://example.org"},
		CheckpointFileJSON: checkpoint,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sp2.Close()
	if len(sp2.snapshotSeen()) == 0 {
		t.Fatal("expected restored checkpoint entries")
	}
}

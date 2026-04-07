package goddgs

import "testing"

func TestFetcherSessionPoolSize(t *testing.T) {
	p := NewFetcherSessionPool()
	if got := p.Size(FetcherKindStealth); got != 0 {
		t.Fatalf("size=%d want 0", got)
	}
	_ = p.Add(FetcherKindStealth, mockFetcher("s1"))
	if got := p.Size(FetcherKindStealth); got != 1 {
		t.Fatalf("size=%d want 1", got)
	}
}

package goddgs

import (
	"context"
	"testing"
)

type mockFetcher string

func (m mockFetcher) Fetch(_ context.Context, _ FetchRequest) (*FetchResponse, error) {
	return &FetchResponse{Fetcher: string(m), StatusCode: 200}, nil
}

func TestFetcherSessionPoolRoundRobin(t *testing.T) {
	p := NewFetcherSessionPool()
	if err := p.Add(FetcherKindHTTP, mockFetcher("a")); err != nil {
		t.Fatal(err)
	}
	if err := p.Add(FetcherKindHTTP, mockFetcher("b")); err != nil {
		t.Fatal(err)
	}
	f1, _ := p.Acquire(FetcherKindHTTP)
	r1, _ := f1.Fetch(context.Background(), FetchRequest{})
	f2, _ := p.Acquire(FetcherKindHTTP)
	r2, _ := f2.Fetch(context.Background(), FetchRequest{})
	if r1.Fetcher == r2.Fetcher {
		t.Fatalf("expected round-robin, got %q then %q", r1.Fetcher, r2.Fetcher)
	}
}

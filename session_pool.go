package goddgs

import (
	"fmt"
	"sync"
)

// FetcherKind identifies a fetcher family.
type FetcherKind string

const (
	FetcherKindHTTP    FetcherKind = "http"
	FetcherKindStealth FetcherKind = "stealth"
	FetcherKindDynamic FetcherKind = "dynamic"
)

// FetcherSessionPool maintains pooled fetchers for hybrid routing.
type FetcherSessionPool struct {
	mu sync.Mutex

	pools   map[FetcherKind][]Fetcher
	nextIdx map[FetcherKind]int
}

func NewFetcherSessionPool() *FetcherSessionPool {
	return &FetcherSessionPool{
		pools:   map[FetcherKind][]Fetcher{},
		nextIdx: map[FetcherKind]int{},
	}
}

func (p *FetcherSessionPool) Add(kind FetcherKind, fetcher Fetcher) error {
	if fetcher == nil {
		return fmt.Errorf("fetcher is nil")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pools[kind] = append(p.pools[kind], fetcher)
	return nil
}

func (p *FetcherSessionPool) Acquire(kind FetcherKind) (Fetcher, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	pool := p.pools[kind]
	if len(pool) == 0 {
		return nil, fmt.Errorf("no fetchers registered for kind %q", kind)
	}
	i := p.nextIdx[kind] % len(pool)
	p.nextIdx[kind] = (i + 1) % len(pool)
	return pool[i], nil
}

func (p *FetcherSessionPool) Size(kind FetcherKind) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.pools[kind])
}

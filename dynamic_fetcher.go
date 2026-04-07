package goddgs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DynamicFetchRequest extends FetchRequest with browser-flow controls.
type DynamicFetchRequest struct {
	FetchRequest
	WaitForSelector string
	NetworkIdleWait time.Duration
}

// DynamicFetcher is a higher-level browser fetcher for interactive flows.
// It currently composes StealthyFetcher and adds wait controls/fallback policy.
type DynamicFetcher struct {
	stealth *StealthyFetcher
	opts    StealthOptions
}

func NewDynamicFetcher(opts StealthOptions) (*DynamicFetcher, error) {
	sf, err := NewStealthyFetcher(opts)
	if err != nil {
		return nil, err
	}
	return &DynamicFetcher{stealth: sf, opts: opts.withDefaults()}, nil
}

func (d *DynamicFetcher) Fetch(ctx context.Context, req FetchRequest) (*FetchResponse, error) {
	return d.FetchDynamic(ctx, DynamicFetchRequest{FetchRequest: req})
}

func (d *DynamicFetcher) FetchDynamic(ctx context.Context, req DynamicFetchRequest) (*FetchResponse, error) {
	if strings.TrimSpace(req.URL) == "" {
		return nil, fmt.Errorf("fetch url is required")
	}
	if req.NetworkIdleWait > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(req.NetworkIdleWait):
		}
	}
	// Current implementation delegates to StealthyFetcher; selector waits are
	// expected to be added in browser-page mode as the next evolution step.
	return d.stealth.Fetch(ctx, req.FetchRequest)
}

func (d *DynamicFetcher) Close() error {
	if d.stealth == nil {
		return nil
	}
	return d.stealth.Close()
}

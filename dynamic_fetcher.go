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
	Actions         []DynamicAction
}

// DynamicAction describes one high-level browser interaction step.
type DynamicAction struct {
	Type     string // wait|click|eval
	Selector string
	Script   string
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
	resp, err := d.stealth.Fetch(ctx, req.FetchRequest)
	if err != nil {
		return nil, err
	}
	if req.WaitForSelector != "" && !strings.Contains(string(resp.Body), req.WaitForSelector) {
		return nil, fmt.Errorf("wait selector not found: %s", req.WaitForSelector)
	}
	for _, action := range req.Actions {
		typ := strings.ToLower(strings.TrimSpace(action.Type))
		switch typ {
		case "wait", "click":
			if strings.TrimSpace(action.Selector) == "" {
				return nil, fmt.Errorf("%s action requires selector", typ)
			}
			if !strings.Contains(string(resp.Body), action.Selector) {
				return nil, fmt.Errorf("%s selector not found: %s", typ, action.Selector)
			}
		case "eval":
			if strings.TrimSpace(action.Script) == "" {
				return nil, fmt.Errorf("eval action requires script")
			}
		default:
			return nil, fmt.Errorf("unsupported dynamic action: %s", action.Type)
		}
	}
	return resp, nil
}

func (d *DynamicFetcher) Close() error {
	if d.stealth == nil {
		return nil
	}
	return d.stealth.Close()
}

package goddgs

import (
	"net/http"
	"time"
)

// StealthLevel controls how much browser hardening is applied.
type StealthLevel string

const (
	StealthLevelBasic      StealthLevel = "basic"
	StealthLevelStrong     StealthLevel = "strong"
	StealthLevelAggressive StealthLevel = "aggressive"
)

// StealthOptions controls both HTTP and browser-backed fetchers.
//
// AntiBotConfig is embedded so all existing anti-bot primitives are inherited
// by default (UA rotation, ChromeTLS, adaptive pacing, circuit breaker,
// challenge solver chain, and proxy pool).
type StealthOptions struct {
	*AntiBotConfig

	HTTPClient *http.Client

	// Browser options.
	Headless             bool
	HumanLikeBehavior    bool
	StealthLevel         StealthLevel
	PersistentContextDir string
	BrowserBinary        string
	ProxyURL             string

	// Request shaping.
	UserAgent    string
	ExtraHeaders map[string]string

	// Operational limits.
	RequestTimeout time.Duration
	MaxBodyBytes   int64
}

func (o StealthOptions) withDefaults() StealthOptions {
	if o.AntiBotConfig == nil {
		o.AntiBotConfig = NewAntiBotConfig()
	}
	if o.RequestTimeout <= 0 {
		o.RequestTimeout = 20 * time.Second
	}
	if o.MaxBodyBytes <= 0 {
		o.MaxBodyBytes = 4 << 20 // 4 MiB
	}
	if o.StealthLevel == "" {
		o.StealthLevel = StealthLevelStrong
	}
	if !o.Headless {
		// Keep default true unless explicitly set false by caller.
		o.Headless = true
	}
	if o.ExtraHeaders == nil {
		o.ExtraHeaders = map[string]string{}
	}
	return o
}

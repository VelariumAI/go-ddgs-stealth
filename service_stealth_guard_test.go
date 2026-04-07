package goddgs

import (
	"net/http"
	"testing"
	"time"
)

func TestStealthGuardHelpers(t *testing.T) {
	resetStealthGuardForTests()
	t.Setenv("GODDGS_STEALTH_RATE_PER_MIN", "2")
	firstAllowed := stealthGuard.allow("1.2.3.4")
	secondAllowed := stealthGuard.allow("1.2.3.4")
	if !firstAllowed || !secondAllowed {
		t.Fatal("expected first two allows")
	}
	if stealthGuard.allow("1.2.3.4") {
		t.Fatal("expected third deny")
	}

	t.Setenv("GODDGS_STEALTH_RATE_PER_MIN", "bad")
	if n := stealthRateLimitPerMin(); n != 120 {
		t.Fatalf("rate default=%d", n)
	}
	t.Setenv("GODDGS_STEALTH_RATE_PER_MIN", "0")
	if n := stealthRateLimitPerMin(); n != 0 {
		t.Fatalf("rate zero=%d", n)
	}
	t.Setenv("GODDGS_STEALTH_RATE_PER_MIN", "-1")
	if n := stealthRateLimitPerMin(); n != 120 {
		t.Fatalf("rate negative default=%d", n)
	}

	t.Setenv("GODDGS_API_TOKEN", "secret")
	r, _ := http.NewRequest(http.MethodGet, "http://x", nil)
	r.Header.Set("Authorization", "Bearer secret")
	if !requireAPIToken(r) {
		t.Fatal("expected valid token")
	}
	r.Header.Set("Authorization", "wrong")
	if requireAPIToken(r) {
		t.Fatal("expected invalid token")
	}

	r.Header.Set("X-Forwarded-For", "9.9.9.9, 8.8.8.8")
	if ip := requesterIP(r); ip != "9.9.9.9" {
		t.Fatalf("xff ip=%s", ip)
	}
	r.Header.Del("X-Forwarded-For")
	r.RemoteAddr = "10.0.0.1:9999"
	if ip := requesterIP(r); ip != "10.0.0.1" {
		t.Fatalf("remote ip=%s", ip)
	}
	r.RemoteAddr = "bad-addr"
	if ip := requesterIP(r); ip != "bad-addr" {
		t.Fatalf("fallback ip=%s", ip)
	}
	r.RemoteAddr = ""
	if ip := requesterIP(r); ip != "unknown" {
		t.Fatalf("unknown ip=%s", ip)
	}
	t.Setenv("GODDGS_API_TOKEN", "")
	if !requireAPIToken(r) {
		t.Fatal("expected open access when token unset")
	}

	stealthGuard.mu.Lock()
	stealthGuard.windowStart = time.Now().Add(-2 * time.Minute)
	stealthGuard.mu.Unlock()
	t.Setenv("GODDGS_STEALTH_RATE_PER_MIN", "1")
	if !stealthGuard.allow("window-reset") {
		t.Fatal("expected allow after window reset")
	}
	t.Setenv("GODDGS_STEALTH_RATE_PER_MIN", "0")
	noLimitFirst := stealthGuard.allow("no-limit")
	noLimitSecond := stealthGuard.allow("no-limit")
	if !noLimitFirst || !noLimitSecond {
		t.Fatal("expected always-allow when limit is disabled")
	}
}

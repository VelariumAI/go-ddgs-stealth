package goddgs

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var stealthGuard = &stealthAccessGuard{counts: map[string]int{}, windowStart: time.Now()}

type stealthAccessGuard struct {
	mu sync.Mutex

	windowStart time.Time
	counts      map[string]int
}

func (g *stealthAccessGuard) allow(ip string) bool {
	limit := stealthRateLimitPerMin()
	if limit <= 0 {
		return true
	}
	now := time.Now()
	g.mu.Lock()
	defer g.mu.Unlock()
	if now.Sub(g.windowStart) >= time.Minute {
		g.windowStart = now
		g.counts = map[string]int{}
	}
	g.counts[ip]++
	return g.counts[ip] <= limit
}

func stealthRateLimitPerMin() int {
	v := strings.TrimSpace(os.Getenv("GODDGS_STEALTH_RATE_PER_MIN"))
	if v == "" {
		return 120
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 120
	}
	return n
}

func requireAPIToken(r *http.Request) bool {
	token := strings.TrimSpace(os.Getenv("GODDGS_API_TOKEN"))
	if token == "" {
		return true
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		auth = strings.TrimSpace(auth[7:])
	}
	return auth == token
}

func requesterIP(r *http.Request) string {
	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	if strings.TrimSpace(r.RemoteAddr) != "" {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return "unknown"
}

func resetStealthGuardForTests() {
	stealthGuard.mu.Lock()
	defer stealthGuard.mu.Unlock()
	stealthGuard.windowStart = time.Now()
	stealthGuard.counts = map[string]int{}
}

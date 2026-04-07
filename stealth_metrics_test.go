package goddgs

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestStealthMetricsObserve(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewStealthMetrics(reg)
	m.BrowserInstance(1)
	m.ObserveFetch("http", 200, 50*time.Millisecond)
	m.ObserveFetch("stealth", 503, 80*time.Millisecond)
	m.ObserveBlock("cf_mitigated")
	m.ObserveAdaptation()

	if _, err := reg.Gather(); err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
}

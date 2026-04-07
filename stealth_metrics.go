package goddgs

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// StealthMetrics tracks browser fetch behavior and adaptation events.
type StealthMetrics struct {
	browserInstances prometheus.Gauge
	fetchTotal       *prometheus.CounterVec
	fetchDuration    *prometheus.HistogramVec
	blockTotal       *prometheus.CounterVec
	adaptEvents      prometheus.Counter
}

func NewStealthMetrics(reg prometheus.Registerer) *StealthMetrics {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	m := &StealthMetrics{
		browserInstances: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "goddgs",
			Subsystem: "stealth",
			Name:      "browser_instances",
			Help:      "Number of live browser instances.",
		}),
		fetchTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "goddgs",
			Subsystem: "stealth",
			Name:      "fetch_total",
			Help:      "Total fetch attempts by fetcher and status.",
		}, []string{"fetcher", "status"}),
		fetchDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goddgs",
			Subsystem: "stealth",
			Name:      "fetch_duration_seconds",
			Help:      "Fetch latency by fetcher.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"fetcher"}),
		blockTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "goddgs",
			Subsystem: "stealth",
			Name:      "block_total",
			Help:      "Detected block/challenge events.",
		}, []string{"detector"}),
		adaptEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "goddgs",
			Subsystem: "stealth",
			Name:      "parse_adaptation_total",
			Help:      "Number of adaptive selector self-heal events.",
		}),
	}
	reg.MustRegister(m.browserInstances, m.fetchTotal, m.fetchDuration, m.blockTotal, m.adaptEvents)
	return m
}

func (m *StealthMetrics) BrowserInstance(delta float64) {
	if m != nil {
		m.browserInstances.Add(delta)
	}
}

func (m *StealthMetrics) ObserveFetch(fetcher string, statusCode int, duration time.Duration) {
	if m == nil {
		return
	}
	status := "ok"
	if statusCode >= 400 {
		status = "error"
	}
	m.fetchTotal.WithLabelValues(fetcher, status).Inc()
	m.fetchDuration.WithLabelValues(fetcher).Observe(duration.Seconds())
}

func (m *StealthMetrics) ObserveBlock(detector string) {
	if m != nil {
		m.blockTotal.WithLabelValues(detector).Inc()
	}
}

func (m *StealthMetrics) ObserveAdaptation() {
	if m != nil {
		m.adaptEvents.Inc()
	}
}

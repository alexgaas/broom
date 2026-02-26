package metrics

import (
	"time"

	"cache/internal/bch"

	"github.com/alexgaas/metrics"
	"github.com/alexgaas/metrics/prometheus"
)

var defaultDurationBuckets = metrics.NewDurationBuckets(
	5*time.Millisecond,
	10*time.Millisecond,
	25*time.Millisecond,
	50*time.Millisecond,
	100*time.Millisecond,
	250*time.Millisecond,
	500*time.Millisecond,
	1*time.Second,
	5*time.Second,
)

type Metrics struct {
	Registry *prometheus.Registry

	// HTTP metrics
	HTTPRequestsTotal       metrics.CounterVec
	HTTPRequestDuration     metrics.TimerVec
	HTTPResponseStatusTotal metrics.CounterVec

	// Bloom filter operation metrics
	BloomAddsTotal         metrics.Counter
	BloomChecksTotal       metrics.Counter
	BloomHitsTotal         metrics.Counter
	BloomMissesTotal       metrics.Counter
	BloomClearsTotal       metrics.Counter
	BloomKeysAddedTotal    metrics.Counter
	BloomOperationDuration metrics.TimerVec
}

func New(prefix string) *Metrics {
	reg := prometheus.NewRegistry(
		prometheus.NewRegistryOpts().SetPrefix(prefix),
	)

	m := &Metrics{
		Registry: reg,

		HTTPRequestsTotal:       reg.CounterVec("http_requests_total", []string{"method", "path", "status"}),
		HTTPRequestDuration:     reg.DurationHistogramVec("http_request_duration_seconds", defaultDurationBuckets, []string{"method", "path"}),
		HTTPResponseStatusTotal: reg.CounterVec("http_response_status_total", []string{"status_class"}),

		BloomAddsTotal:         reg.Counter("bloom_adds_total"),
		BloomChecksTotal:       reg.Counter("bloom_checks_total"),
		BloomHitsTotal:         reg.Counter("bloom_hits_total"),
		BloomMissesTotal:       reg.Counter("bloom_misses_total"),
		BloomClearsTotal:       reg.Counter("bloom_clears_total"),
		BloomKeysAddedTotal:    reg.Counter("bloom_keys_added_total"),
		BloomOperationDuration: reg.DurationHistogramVec("bloom_operation_duration_seconds", defaultDurationBuckets, []string{"operation"}),
	}

	return m
}

func (m *Metrics) RegisterBloomGauges(bc *bch.BloomCache) {
	m.Registry.FuncGauge("bloom_bits_set", func() float64 {
		return float64(bc.Stats().BitsSet)
	})
	m.Registry.FuncGauge("bloom_estimated_item_count", func() float64 {
		return bc.Stats().EstimatedItemCount
	})
}

package metrics

import "github.com/prometheus/client_golang/prometheus"

// Metrics holds all Prometheus metrics for massivepinger ICMP probing.
// Each instance is scoped to a single identifier (e.g. the exporter instance name)
// and uses per-target label values supplied at observation time.
type Metrics struct {
	identifier string

	icmpDurationSeconds        *prometheus.GaugeVec
	icmpIntervalSeconds        *prometheus.GaugeVec
	icmpTimeoutSeconds         *prometheus.GaugeVec
	icmpAverageDurationSeconds *prometheus.GaugeVec
	icmpSent                   *prometheus.HistogramVec
	icmpReceived               *prometheus.HistogramVec
}

var rttBuckets = []float64{0.0001, 0.001, 0.005, 0.010, 0.025, 0.050, 0.100, 0.200, 1.0, 5.0}

// New creates a Metrics instance and registers all metrics against reg.
// identifier is a fixed label value applied to every metric this instance records.
func New(identifier string, reg prometheus.Registerer) *Metrics {
	labelNames := []string{"identifier", "target"}

	m := &Metrics{
		identifier: identifier,

		icmpDurationSeconds: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "icmp_duration_seconds",
			Help: "Latest ICMP round-trip time in seconds.",
		}, labelNames),

		icmpIntervalSeconds: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "icmp_interval_seconds",
			Help: "Configured probe interval in seconds for the target.",
		}, labelNames),

		icmpTimeoutSeconds: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "icmp_timeout_seconds",
			Help: "Configured probe timeout in seconds for the target.",
		}, labelNames),

		icmpAverageDurationSeconds: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "icmp_average_duration_seconds",
			Help: "Rolling average ICMP round-trip time in seconds since last scrape.",
		}, labelNames),

		icmpSent: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "icmp_sent",
			Help:    "Histogram of ICMP packets sent, bucketed by RTT.",
			Buckets: rttBuckets,
		}, labelNames),

		icmpReceived: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "icmp_received",
			Help:    "Histogram of ICMP packets received, bucketed by RTT.",
			Buckets: rttBuckets,
		}, labelNames),
	}

	reg.MustRegister(
		m.icmpDurationSeconds,
		m.icmpIntervalSeconds,
		m.icmpTimeoutSeconds,
		m.icmpAverageDurationSeconds,
		m.icmpSent,
		m.icmpReceived,
	)

	return m
}

// labels builds the standard label map for a given target.
func (m *Metrics) labels(target string) prometheus.Labels {
	return prometheus.Labels{
		"identifier": m.identifier,
		"target":     target,
	}
}

// InitTarget pre-populates configuration gauges and zeroes out runtime gauges
// for target. Call this once per configured target before probing begins.
func (m *Metrics) InitTarget(target string, interval, timeout float64) {
	l := m.labels(target)
	m.icmpIntervalSeconds.With(l).Set(interval)
	m.icmpTimeoutSeconds.With(l).Set(timeout)
	m.icmpDurationSeconds.With(l).Set(0)
	m.icmpAverageDurationSeconds.With(l).Set(0)
}

// ObserveSent records a sent ICMP probe in the icmp_sent histogram.
func (m *Metrics) ObserveSent(target string, rtt float64) {
	m.icmpSent.With(m.labels(target)).Observe(rtt)
}

// ObserveReceived records a received ICMP probe in the icmp_received histogram
// and updates the average duration gauge with the latest RTT value.
func (m *Metrics) ObserveReceived(target string, rtt float64) {
	l := m.labels(target)
	m.icmpReceived.With(l).Observe(rtt)
	m.icmpAverageDurationSeconds.With(l).Set(rtt)
}

// SetLatest updates the icmp_duration_seconds gauge to the most recent RTT.
func (m *Metrics) SetLatest(target string, rtt float64) {
	m.icmpDurationSeconds.With(m.labels(target)).Set(rtt)
}

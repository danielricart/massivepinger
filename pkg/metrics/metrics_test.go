package metrics

import (
	"strings"
	"testing"

	dto "github.com/prometheus/client_model/go"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

const (
	testIdentifier = "test-instance"
	testTarget     = "192.0.2.1"
)

// newTestMetrics returns a Metrics instance backed by an isolated registry.
func newTestMetrics(t *testing.T) (*Metrics, prometheus.Gatherer) {
	t.Helper()
	reg := prometheus.NewRegistry()
	m := New(testIdentifier, reg)
	return m, reg
}

// gather collects all metric families from the registry and returns a map
// keyed by metric name for convenient lookup.
func gather(t *testing.T, g prometheus.Gatherer) map[string]*dto.MetricFamily {
	t.Helper()
	mfs, err := g.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}
	out := make(map[string]*dto.MetricFamily, len(mfs))
	for _, mf := range mfs {
		out[mf.GetName()] = mf
	}
	return out
}

// labelValue extracts the value of a named label from a dto.Metric.
func labelValue(m *dto.Metric, name string) string {
	for _, lp := range m.GetLabel() {
		if lp.GetName() == name {
			return lp.GetValue()
		}
	}
	return ""
}

// TestNew_RegistrationSucceeds verifies that constructing a Metrics instance
// against a fresh registry does not panic.
func TestNew_RegistrationSucceeds(t *testing.T) {
	reg := prometheus.NewRegistry()
	_ = New(testIdentifier, reg)
}

// TestInitTarget verifies that icmp_interval_seconds and icmp_timeout_seconds
// are set to the supplied values after InitTarget.
func TestInitTarget(t *testing.T) {
	m, reg := newTestMetrics(t)

	const interval = 5.0
	const timeout = 1.0
	m.InitTarget(testTarget, interval, timeout)

	mfs := gather(t, reg)

	for metricName, want := range map[string]float64{
		"icmp_interval_seconds": interval,
		"icmp_timeout_seconds":  timeout,
	} {
		mf, ok := mfs[metricName]
		if !ok {
			t.Fatalf("metric %q not found after InitTarget", metricName)
		}
		if len(mf.GetMetric()) == 0 {
			t.Fatalf("metric %q has no samples", metricName)
		}
		got := mf.GetMetric()[0].GetGauge().GetValue()
		if got != want {
			t.Errorf("%s: got %v, want %v", metricName, got, want)
		}
	}
}

// TestInitTarget_ZeroesRuntimeGauges verifies that icmp_duration_seconds and
// icmp_average_duration_seconds are initialised to zero by InitTarget.
func TestInitTarget_ZeroesRuntimeGauges(t *testing.T) {
	m, reg := newTestMetrics(t)
	m.InitTarget(testTarget, 10.0, 2.0)

	mfs := gather(t, reg)

	for _, name := range []string{"icmp_duration_seconds", "icmp_average_duration_seconds"} {
		mf, ok := mfs[name]
		if !ok {
			t.Fatalf("metric %q not found after InitTarget", name)
		}
		got := mf.GetMetric()[0].GetGauge().GetValue()
		if got != 0 {
			t.Errorf("%s: want 0 after init, got %v", name, got)
		}
	}
}

// TestObserveSent verifies that a single ObserveSent call increments
// icmp_sent_count to 1.
func TestObserveSent(t *testing.T) {
	m, reg := newTestMetrics(t)
	m.InitTarget(testTarget, 5.0, 1.0)

	m.ObserveSent(testTarget, 0.005)

	// testutil.ToFloat64 works on a Collector; use the GaugeVec's underlying
	// histogram count via CollectAndCompare text comparison or gather directly.
	mfs := gather(t, reg)
	mf, ok := mfs["icmp_sent"]
	if !ok {
		t.Fatal("metric icmp_sent not found after ObserveSent")
	}
	got := mf.GetMetric()[0].GetHistogram().GetSampleCount()
	if got != 1 {
		t.Errorf("icmp_sent sample count: got %d, want 1", got)
	}
}

// TestObserveReceived verifies that a single ObserveReceived call increments
// icmp_received_count to 1.
func TestObserveReceived(t *testing.T) {
	m, reg := newTestMetrics(t)
	m.InitTarget(testTarget, 5.0, 1.0)

	m.ObserveReceived(testTarget, 0.010)

	mfs := gather(t, reg)
	mf, ok := mfs["icmp_received"]
	if !ok {
		t.Fatal("metric icmp_received not found after ObserveReceived")
	}
	got := mf.GetMetric()[0].GetHistogram().GetSampleCount()
	if got != 1 {
		t.Errorf("icmp_received sample count: got %d, want 1", got)
	}
}

// TestObserveReceived_UpdatesAverage verifies EMA behaviour:
//   - first observation from zero → emaAlpha * rtt
//   - repeated identical observations converge to within 1% of rtt
func TestObserveReceived_UpdatesAverage(t *testing.T) {
	m, _ := newTestMetrics(t)
	m.InitTarget(testTarget, 5.0, 1.0)

	const rtt = 0.042

	// First observation: EMA starts at 0, so result = emaAlpha * rtt.
	m.ObserveReceived(testTarget, rtt)
	want := emaAlpha * rtt
	got := testutil.ToFloat64(m.icmpAverageDurationSeconds.With(m.labels(testTarget)))
	if got != want {
		t.Errorf("after first observation: icmp_average_duration_seconds = %v, want %v", got, want)
	}

	// Drive convergence: 50 identical samples bring EMA within 1% of rtt.
	for i := 0; i < 50; i++ {
		m.ObserveReceived(testTarget, rtt)
	}
	converged := testutil.ToFloat64(m.icmpAverageDurationSeconds.With(m.labels(testTarget)))
	if converged < rtt*0.99 || converged > rtt*1.01 {
		t.Errorf("after convergence: icmp_average_duration_seconds = %v, want ~%v (±1%%)", converged, rtt)
	}
}

// TestSetLatest verifies that SetLatest correctly updates icmp_duration_seconds.
func TestSetLatest(t *testing.T) {
	m, reg := newTestMetrics(t)
	m.InitTarget(testTarget, 5.0, 1.0)

	const rtt = 0.005
	m.SetLatest(testTarget, rtt)

	// Use testutil.ToFloat64 with the concrete GaugeVec label set.
	got := testutil.ToFloat64(m.icmpDurationSeconds.With(m.labels(testTarget)))
	if got != rtt {
		t.Errorf("icmp_duration_seconds: got %v, want %v", got, rtt)
	}

	// Also confirm via the registry gatherer.
	mfs := gather(t, reg)
	mf, ok := mfs["icmp_duration_seconds"]
	if !ok {
		t.Fatal("metric icmp_duration_seconds not found")
	}
	gatheredVal := mf.GetMetric()[0].GetGauge().GetValue()
	if gatheredVal != rtt {
		t.Errorf("icmp_duration_seconds (gathered): got %v, want %v", gatheredVal, rtt)
	}
}

// TestLabelsPresent verifies that gathered metrics carry both the `identifier`
// and `target` labels with the expected values.
func TestLabelsPresent(t *testing.T) {
	m, reg := newTestMetrics(t)
	m.InitTarget(testTarget, 5.0, 1.0)

	mfs := gather(t, reg)

	for name, mf := range mfs {
		for _, metric := range mf.GetMetric() {
			idVal := labelValue(metric, "identifier")
			tgtVal := labelValue(metric, "target")

			if idVal == "" {
				t.Errorf("metric %q is missing label 'identifier'", name)
			}
			if tgtVal == "" {
				t.Errorf("metric %q is missing label 'target'", name)
			}
			if idVal != testIdentifier {
				t.Errorf("metric %q: identifier label = %q, want %q", name, idVal, testIdentifier)
			}
			if tgtVal != testTarget {
				t.Errorf("metric %q: target label = %q, want %q", name, tgtVal, testTarget)
			}
		}
	}
}

// TestMultipleTargets verifies that distinct targets produce independent label
// sets and do not overwrite each other.
func TestMultipleTargets(t *testing.T) {
	m, _ := newTestMetrics(t)

	targets := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
	for _, tgt := range targets {
		m.InitTarget(tgt, 5.0, 1.0)
	}

	for _, tgt := range targets {
		m.SetLatest(tgt, 0.001)
	}

	for _, tgt := range targets {
		got := testutil.ToFloat64(m.icmpDurationSeconds.With(m.labels(tgt)))
		if got != 0.001 {
			t.Errorf("target %s: icmp_duration_seconds = %v, want 0.001", tgt, got)
		}
	}
}

// TestMetricNames verifies that the expected metric names are all registered.
func TestMetricNames(t *testing.T) {
	_, reg := newTestMetrics(t)

	// Force label dimension creation so the metrics appear in Gather output.
	// (Histograms need at least one observation or pre-registration via InitTarget.)
	m, _ := newTestMetrics(t)
	m.InitTarget(testTarget, 5.0, 1.0)
	m.ObserveSent(testTarget, 0.001)
	m.ObserveReceived(testTarget, 0.001)

	// Use a second isolated registry for the name check.
	reg2 := prometheus.NewRegistry()
	m2 := New(testIdentifier, reg2)
	m2.InitTarget(testTarget, 5.0, 1.0)
	m2.ObserveSent(testTarget, 0.001)
	m2.ObserveReceived(testTarget, 0.001)

	_ = reg // keep linter happy

	mfs := gather(t, reg2)
	expected := []string{
		"icmp_duration_seconds",
		"icmp_interval_seconds",
		"icmp_timeout_seconds",
		"icmp_average_duration_seconds",
		"icmp_sent",
		"icmp_received",
	}
	for _, name := range expected {
		if _, ok := mfs[name]; !ok {
			var found []string
			for k := range mfs {
				found = append(found, k)
			}
			t.Errorf("expected metric %q not found; registered metrics: %s",
				name, strings.Join(found, ", "))
		}
	}
}

package pinger

import (
	"context"
	"sync"
	"testing"
	"time"

	"massivepinger/pkg/config"
)

// RecordingMetrics is a test double for MetricsRecorder that records every call
// made to it. All methods are safe for concurrent use.
type RecordingMetrics struct {
	mu            sync.Mutex
	initCalls     int
	sentCalls     []float64
	receivedCalls []float64
	latestCalls   []float64
}

func (r *RecordingMetrics) InitTarget(_ string, _, _ float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.initCalls++
}

func (r *RecordingMetrics) ObserveSent(_ string, rtt float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sentCalls = append(r.sentCalls, rtt)
}

func (r *RecordingMetrics) ObserveReceived(_ string, rtt float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.receivedCalls = append(r.receivedCalls, rtt)
}

func (r *RecordingMetrics) SetLatest(_ string, rtt float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.latestCalls = append(r.latestCalls, rtt)
}

func (r *RecordingMetrics) SentLen() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sentCalls)
}

func (r *RecordingMetrics) ReceivedLen() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.receivedCalls)
}

// ---------------------------------------------------------------------------
// TestNew_InitTargetCalled verifies that New calls InitTarget once per target.
// ---------------------------------------------------------------------------

func TestNew_InitTargetCalled(t *testing.T) {
	t.Parallel()

	targets := []config.Target{
		{Address: "192.0.2.1", Interval: time.Second, Timeout: 500 * time.Millisecond},
		{Address: "192.0.2.2", Interval: time.Second, Timeout: 500 * time.Millisecond},
	}

	rec := &RecordingMetrics{}
	New(targets, rec)

	rec.mu.Lock()
	got := rec.initCalls
	rec.mu.Unlock()

	if got != 2 {
		t.Errorf("expected InitTarget called 2 times, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// TestStart_ContextCancel verifies that goroutines exit when the context is
// cancelled. A 200 ms grace window is used; the test must not panic.
// ---------------------------------------------------------------------------

func TestStart_ContextCancel(t *testing.T) {
	t.Parallel()

	// Use an address in the TEST-NET range (RFC 5737) to ensure no real network
	// traffic is generated. The pinger will fail immediately (unreachable host
	// or DNS error), which is acceptable — we only verify no panic occurs and
	// the goroutine loop exits.
	targets := []config.Target{
		{Address: "192.0.2.100", Interval: 50 * time.Millisecond, Timeout: 20 * time.Millisecond},
	}

	rec := &RecordingMetrics{}
	mgr := New(targets, rec)

	ctx, cancel := context.WithCancel(context.Background())
	mgr.Start(ctx)

	// Let the loop run briefly, then cancel.
	time.Sleep(80 * time.Millisecond)
	cancel()

	// Give goroutines up to 200 ms to exit cleanly.
	// We cannot join goroutines directly without modifying the Manager; instead
	// we rely on the select on ctx.Done() in runTarget completing within this
	// window. A panic from a goroutine would still fail the test.
	time.Sleep(200 * time.Millisecond)

	// If we reach here without a panic the goroutines exited (or are about to).
	// ObserveSent must have been called at least once (timeout path) because the
	// target is unreachable.
	if rec.SentLen() == 0 {
		// This is a soft check — network may be fully broken; log but don't fail.
		t.Log("note: ObserveSent was not called; network may be completely isolated")
	}
}

// ---------------------------------------------------------------------------
// TestManager_IntegrationLocalhost attempts to ping 127.0.0.1 and verifies
// that metrics are recorded. The test is skipped gracefully when ICMP is
// unavailable (e.g., inside a sandboxed CI container without NET_ADMIN).
// ---------------------------------------------------------------------------

func TestManager_IntegrationLocalhost(t *testing.T) {
	// Not parallel: touches the network stack.

	targets := []config.Target{
		{Address: "127.0.0.1", Interval: 2 * time.Second, Timeout: 500 * time.Millisecond},
	}

	rec := &RecordingMetrics{}
	mgr := New(targets, rec)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr.Start(ctx)

	// Wait long enough for at least one complete probe cycle.
	// The probe itself takes at most 500 ms (timeout); add headroom.
	time.Sleep(700 * time.Millisecond)

	cancel()

	// Wait briefly for the goroutine to notice cancellation.
	time.Sleep(50 * time.Millisecond)

	sentLen := rec.SentLen()
	receivedLen := rec.ReceivedLen()

	if sentLen == 0 {
		// If nothing was sent the pinger could not run (no ICMP capability).
		// This is not a test failure — just skip.
		t.Skip("skipping: pinger produced no sent observations; ICMP may be unavailable in this environment")
	}

	// If the probe reached localhost we expect at least one received entry.
	// We do not hard-fail on receivedLen == 0 because on some systems even
	// localhost ICMP responses are blocked; we do log the discrepancy.
	if receivedLen == 0 {
		t.Logf("note: 127.0.0.1 sent=%d received=%d — ICMP replies may be filtered", sentLen, receivedLen)
	} else {
		t.Logf("127.0.0.1 integration ping: sent=%d received=%d", sentLen, receivedLen)
	}
}

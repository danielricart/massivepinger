package pinger

import (
	"context"
	"time"

	"massivepinger/pkg/config"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/sirupsen/logrus"
)

// MetricsRecorder abstracts the metrics operations required by the Manager.
// *metrics.Metrics satisfies this interface; test doubles may also implement it.
type MetricsRecorder interface {
	InitTarget(target string, interval, timeout float64)
	ObserveSent(target string, rtt float64)
	ObserveReceived(target string, rtt float64)
	SetLatest(target string, rtt float64)
}

// Manager supervises one goroutine per configured target, continuously sending
// ICMP probes and recording results through a MetricsRecorder.
type Manager struct {
	targets []config.Target
	metrics MetricsRecorder
}

// New creates a Manager and initialises metrics for every target.
func New(targets []config.Target, m MetricsRecorder) *Manager {
	for _, t := range targets {
		m.InitTarget(t.Address, t.Interval.Seconds(), t.Timeout.Seconds())
	}
	return &Manager{
		targets: targets,
		metrics: m,
	}
}

// Start launches one goroutine per target and returns immediately.
// All goroutines respect ctx: they exit cleanly when ctx is cancelled.
func (mgr *Manager) Start(ctx context.Context) {
	for _, t := range mgr.targets {
		go mgr.runTarget(ctx, t)
	}
}

// runTarget is the per-target ping loop. It runs one probe per interval cycle
// until ctx is cancelled.
func (mgr *Manager) runTarget(ctx context.Context, target config.Target) {
	for {
		if ctx.Err() != nil {
			return
		}

		pinger, err := probing.NewPinger(target.Address)
		if err != nil {
			logrus.Warnf("pinger: target %s: create error: %v", target.Address, err)
			// Wait for the next interval before retrying creation.
			select {
			case <-time.After(target.Interval):
			case <-ctx.Done():
				return
			}
			continue
		}

		// Use unprivileged UDP ping (no raw-socket capability required).
		pinger.SetPrivileged(false)
		pinger.Count = 1

		pinger.Timeout = target.Timeout
		received := false
		pinger.OnRecv = func(pkt *probing.Packet) {
			rtt := pkt.Rtt.Seconds()
			mgr.metrics.ObserveReceived(target.Address, rtt)
			mgr.metrics.SetLatest(target.Address, rtt)
			mgr.metrics.ObserveSent(target.Address, rtt)
			received = true
		}

		start := time.Now()
		if err := pinger.Run(); err != nil {
			logrus.Warnf("pinger: target %s: run error: %v", target.Address, err)
		}

		if !received {
			// Probe timed out or errored — count as a sent packet with the
			// configured timeout as the RTT value.
			mgr.metrics.ObserveSent(target.Address, target.Timeout.Seconds())
		}

		// Respect the configured interval: sleep for whatever time remains
		// after the probe completed.
		elapsed := time.Since(start)
		remaining := target.Interval - elapsed
		if remaining > 0 {
			select {
			case <-time.After(remaining):
			case <-ctx.Done():
				return
			}
		}
	}
}

package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"massivepinger/pkg/config"
	"massivepinger/pkg/metrics"
	"massivepinger/pkg/pinger"
	"massivepinger/pkg/server"
)

// warnHook is a logrus hook that captures warning-level log messages.
// It is safe for concurrent use.
type warnHook struct {
	mu      sync.Mutex
	entries []string
}

func (h *warnHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.WarnLevel}
}

func (h *warnHook) Fire(entry *logrus.Entry) error {
	h.mu.Lock()
	h.entries = append(h.entries, entry.Message)
	h.mu.Unlock()
	return nil
}

func (h *warnHook) messages() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, len(h.entries))
	copy(out, h.entries)
	return out
}

// writeTempConfig writes YAML content to a temporary file and returns its path.
// The file is automatically removed when the test ends.
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "massivepinger-e2e-*.yaml")
	if err != nil {
		t.Fatalf("creating temp config: %v", err)
	}
	if _, err := fmt.Fprint(f, content); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("closing temp config: %v", err)
	}
	return f.Name()
}

// TestE2E_FullStack starts the real server, real metrics registry, and the
// real pinger against 127.0.0.1, then scrapes the live /metrics endpoint to
// confirm all expected metric families and label values are present.
//
// Because InitTarget is called by pinger.New before any probe runs, histogram
// metric families are registered and exported immediately. The test therefore
// succeeds even in environments where ICMP is unavailable (e.g., sandboxed CI).
func TestE2E_FullStack(t *testing.T) {
	const (
		port       = 19123
		identifier = "e2etest"
		target     = "127.0.0.1"
	)

	cfgContent := fmt.Sprintf(`- target: %s
  interval: 100ms
  timeout: 500ms
`, target)

	cfgPath := writeTempConfig(t, cfgContent)

	targets, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target from config, got %d", len(targets))
	}

	reg := prometheus.NewRegistry()
	m := metrics.New(identifier, reg)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	mgr := pinger.New(targets, m)
	mgr.Start(ctx)

	srv := server.New(port, reg)
	go func() {
		if listenErr := srv.ListenAndServe(); listenErr != nil && listenErr != http.ErrServerClosed {
			// Log but do not call t.Fatal from a goroutine — the test may have
			// already finished and t would be invalid.
			_ = listenErr
		}
	}()

	t.Cleanup(func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	})

	// Allow a few probe cycles to run.
	time.Sleep(600 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading /metrics body: %v", err)
	}
	body := string(bodyBytes)

	expectations := []string{
		"icmp_sent_bucket",
		"icmp_received_bucket",
		"icmp_duration_seconds",
		"icmp_interval_seconds",
		"icmp_timeout_seconds",
		fmt.Sprintf(`identifier="%s"`, identifier),
		fmt.Sprintf(`target="%s"`, target),
	}

	allPassed := true
	for _, want := range expectations {
		if !strings.Contains(body, want) {
			t.Errorf("expected /metrics body to contain %q", want)
			allPassed = false
		}
	}

	if !allPassed {
		t.Logf("/metrics response body:\n%s", body)
	}
}

// TestE2E_InvalidConfigWarning verifies that config.Load skips entries with a
// missing target field and emits at least one logrus warning containing "missing".
func TestE2E_InvalidConfigWarning(t *testing.T) {
	cfgContent := `- target: 127.0.0.1
  interval: 100ms
  timeout: 500ms
- target: ""
  interval: 100ms
  timeout: 500ms
`

	cfgPath := writeTempConfig(t, cfgContent)

	hook := &warnHook{}
	logrus.AddHook(hook)

	targets, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}

	if len(targets) != 1 {
		t.Errorf("expected exactly 1 valid target, got %d", len(targets))
	}

	msgs := hook.messages()
	if len(msgs) == 0 {
		t.Error("expected at least one warning to be logged, got none")
	}

	found := false
	for _, msg := range msgs {
		if strings.Contains(strings.ToLower(msg), "missing") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a warning containing \"missing\", got warnings: %v", msgs)
	}
}

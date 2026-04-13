package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeTemp writes content to a file named name inside t.TempDir() and returns its path.
func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeTemp: %v", err)
	}
	return path
}

func TestLoad_ValidYAML(t *testing.T) {
	yaml := `
- target: 192.168.1.1
  interval: 1s
  timeout: 500ms
- target: 10.0.0.1
  interval: 200ms
  timeout: 2s
`
	path := writeTemp(t, "targets.yaml", yaml)

	targets, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}

	first := targets[0]
	if first.Address != "192.168.1.1" {
		t.Errorf("first.Address: got %q, want %q", first.Address, "192.168.1.1")
	}
	if first.Interval != time.Second {
		t.Errorf("first.Interval: got %v, want %v", first.Interval, time.Second)
	}
	if first.Timeout != 500*time.Millisecond {
		t.Errorf("first.Timeout: got %v, want %v", first.Timeout, 500*time.Millisecond)
	}

	second := targets[1]
	if second.Address != "10.0.0.1" {
		t.Errorf("second.Address: got %q, want %q", second.Address, "10.0.0.1")
	}
	if second.Interval != 200*time.Millisecond {
		t.Errorf("second.Interval: got %v, want %v", second.Interval, 200*time.Millisecond)
	}
	if second.Timeout != 2*time.Second {
		t.Errorf("second.Timeout: got %v, want %v", second.Timeout, 2*time.Second)
	}
}

func TestLoad_InfiniteTimeout(t *testing.T) {
	yaml := `
- target: 8.8.8.8
  interval: 1s
  timeout: infinite
`
	path := writeTemp(t, "targets.yaml", yaml)

	targets, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Timeout != 120*time.Second {
		t.Errorf("expected Timeout == 120s for infinite, got %v", targets[0].Timeout)
	}
}

func TestLoad_MissingTargetField(t *testing.T) {
	yaml := `
- target: ""
  interval: 1s
  timeout: 500ms
`
	path := writeTemp(t, "targets.yaml", yaml)

	targets, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 0 {
		t.Errorf("expected 0 targets (entry skipped), got %d", len(targets))
	}
}

func TestLoad_MissingInterval(t *testing.T) {
	yaml := `
- target: 192.168.0.1
  interval: ""
  timeout: 1s
`
	path := writeTemp(t, "targets.yaml", yaml)

	targets, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 0 {
		t.Errorf("expected 0 targets (entry skipped), got %d", len(targets))
	}
}

func TestLoad_BadDurationFormat(t *testing.T) {
	yaml := `
- target: 192.168.0.1
  interval: notaduration
  timeout: 1s
`
	path := writeTemp(t, "targets.yaml", yaml)

	targets, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 0 {
		t.Errorf("expected 0 targets (entry skipped), got %d", len(targets))
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_MixedValidAndInvalid(t *testing.T) {
	yaml := `
- target: 10.10.10.1
  interval: 1s
  timeout: 500ms
- target: 10.10.10.2
  interval: 1s
  timeout: ""
`
	path := writeTemp(t, "targets.yaml", yaml)

	targets, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Address != "10.10.10.1" {
		t.Errorf("got address %q, want %q", targets[0].Address, "10.10.10.1")
	}
}

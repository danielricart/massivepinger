package config

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Target holds the parsed configuration for a single ping target.
type Target struct {
	Address  string
	Interval time.Duration
	Timeout  time.Duration
}

// rawTarget mirrors the YAML structure before validation and type conversion.
type rawTarget struct {
	Target   string `yaml:"target"`
	Interval string `yaml:"interval"`
	Timeout  string `yaml:"timeout"`
}

// Load reads the YAML file at path and returns the slice of valid Target entries.
// File read or YAML parse failures are returned as errors. Individual entries
// that fail validation are skipped with a logrus warning.
func Load(path string) ([]Target, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: reading file %q: %w", path, err)
	}

	var raws []rawTarget
	if err := yaml.Unmarshal(data, &raws); err != nil {
		return nil, fmt.Errorf("config: parsing YAML from %q: %w", path, err)
	}

	targets := make([]Target, 0, len(raws))

	for i, r := range raws {
		if r.Target == "" {
			logrus.Warnf("config: entry %d: missing or empty 'target' field, skipping", i)
			continue
		}

		if r.Interval == "" {
			logrus.Warnf("config: entry %d (%s): missing or empty 'interval' field, skipping", i, r.Target)
			continue
		}

		interval, err := time.ParseDuration(r.Interval)
		if err != nil {
			logrus.Warnf("config: entry %d (%s): invalid 'interval' value %q: %v, skipping", i, r.Target, r.Interval, err)
			continue
		}

		if r.Timeout == "" {
			logrus.Warnf("config: entry %d (%s): missing or empty 'timeout' field, skipping", i, r.Target)
			continue
		}

		var timeout time.Duration
		if r.Timeout == "infinite" {
			logrus.Warnf("config: entry %d (%s): 'infinite' timeout is not supported; defaulting to 120s", i, r.Target)
			timeout = 120 * time.Second
		} else {
			timeout, err = time.ParseDuration(r.Timeout)
			if err != nil {
				logrus.Warnf("config: entry %d (%s): invalid 'timeout' value %q: %v, skipping", i, r.Target, r.Timeout, err)
				continue
			}
		}

		targets = append(targets, Target{
			Address:  r.Target,
			Interval: interval,
			Timeout:  timeout,
		})
	}

	return targets, nil
}

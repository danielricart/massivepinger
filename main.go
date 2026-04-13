package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"massivepinger/pkg/config"
	"massivepinger/pkg/metrics"
	"massivepinger/pkg/pinger"
	"massivepinger/pkg/server"
)

func runE(cmd *cobra.Command, args []string) error {
	// 1. Parse flags
	cfgPath, _ := cmd.Flags().GetString("config")
	port, _ := cmd.Flags().GetInt("port")
	identifier, _ := cmd.Flags().GetString("identifier")

	// 2. Load config
	targets, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if len(targets) == 0 {
		logrus.Warn("no valid targets found in config, exiting")
		return nil
	}
	logrus.Infof("loaded %d target(s)", len(targets))

	// 3. Create Prometheus registry + metrics
	reg := prometheus.NewRegistry()
	// Also register default Go metrics collectors
	reg.MustRegister(prometheus.NewGoCollector(), prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	m := metrics.New(identifier, reg)

	// 4. Create pinger manager
	mgr := pinger.New(targets, m)

	// 5. Create and start HTTP server
	srv := server.New(port, reg)
	go func() {
		logrus.Infof("metrics server listening on :%d", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("metrics server error: %v", err)
		}
	}()

	// 6. Start pinger manager (non-blocking, spawns goroutines)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr.Start(ctx)
	logrus.Info("pinger started")

	// 7. Wait for signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	logrus.Infof("received signal %v, shutting down", sig)

	// 8. Graceful shutdown
	cancel() // stop pingers
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logrus.Warnf("server shutdown error: %v", err)
	}
	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:          "massivepinger",
		Short:        "Continuous ICMP ping exporter for Prometheus",
		RunE:         runE,
		SilenceUsage: true,
	}
	rootCmd.Flags().String("config", "config.yaml", "path to config file")
	rootCmd.Flags().Int("port", 9123, "metrics HTTP port")
	rootCmd.Flags().String("identifier", "massivepinger", "stable identifier for this instance")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

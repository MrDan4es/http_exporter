package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/pflag"

	"github.com/mrdan4es/http_exporter/pkg/collector"
	"github.com/mrdan4es/http_exporter/pkg/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	configPath := pflag.StringP("config", "c", "", "path to config file")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}

	log := zlog.Output(
		zerolog.NewConsoleWriter(
			func(w *zerolog.ConsoleWriter) {
				w.TimeFormat = time.RFC3339Nano
			},
		),
	)
	ctx = log.WithContext(ctx)

	msg, err := json.Marshal(cfg)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("marshal config")
	}

	log.Info().
		RawJSON("config", msg).
		Send()

	if err := run(ctx, cfg); err != nil {
		log.Fatal().
			Err(err).
			Send()
	}
}

func run(ctx context.Context, cfg *config.Config) error {
	log := zerolog.Ctx(ctx)

	for _, collectorCfg := range cfg.Collectors {
		c, err := collector.New(ctx, collectorCfg)
		if err != nil {
			return fmt.Errorf("create %s collector: %w", collectorCfg.Name, err)
		}

		prometheus.MustRegister(c)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	httpServer := &http.Server{
		Addr:    cfg.HTTP.ListenAddr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()

		log.Info().Msg("Shutting down HTTP server...")

		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Err(err).Send()
		}
	}()

	log.Info().Msgf("Serving /metrics on addr %s...", cfg.HTTP.ListenAddr)

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

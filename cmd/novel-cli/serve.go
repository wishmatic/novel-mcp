package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/wishmatic/novel-mcp/internal/cli"
	"github.com/wishmatic/novel-mcp/internal/server"
)

const shutdownTimeout = 15 * time.Second

func newServeCommand(deps cli.Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the MCP server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return serve(cmd.Context(), deps)
		},
	}
}

func serve(parent context.Context, deps cli.Deps) error {
	cfg, err := cli.LoadConfig(deps)
	if err != nil {
		return err
	}

	log := newLogger(cfg.LogLevel)
	defer log.Sync()

	srv, err := server.New(cfg, log)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Error("server failed", zap.Error(err))

			return err
		}
	case <-ctx.Done():
		log.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			log.Error("graceful shutdown failed", zap.Error(err))
		}
	}

	log.Info("server stopped")

	return nil
}

func newLogger(level string) *zap.Logger {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = zapcore.InfoLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)

	log, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return log
}

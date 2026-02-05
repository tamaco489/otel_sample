package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tamaco489/otel_sample/02_instrumentation/internal/config"
	"github.com/tamaco489/otel_sample/02_instrumentation/internal/di"
	"github.com/tamaco489/otel_sample/02_instrumentation/pkg/library/otel"
)

func main() {
	ctx := context.Background()

	// slog の設定
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// 設定の読み込み
	cfg := config.NewConfig()

	// OTEL Provider の初期化
	provider, err := otel.NewProvider(ctx, otel.Config{
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		Environment:    cfg.Environment,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to initialize otel", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 依存関係の初期化
	container := di.NewContainer()

	// サーバー起動 (別goroutine)
	go func() {
		if err := container.Server.Run(ctx, ":8080"); err != nil {
			slog.ErrorContext(ctx, "server error", slog.String("error", err.Error()))
		}
	}()

	// シグナル待機
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	slog.InfoContext(ctx, "shutting down...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := container.Server.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, "failed to shutdown server", slog.String("error", err.Error()))
	}
	if err := provider.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(shutdownCtx, "failed to shutdown otel", slog.String("error", err.Error()))
	}

	slog.InfoContext(ctx, "shutdown complete")
}

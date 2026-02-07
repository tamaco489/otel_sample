// pkg/otel/provider.go

package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config は OTEL 初期化の設定
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
}

// Provider は OTEL の各種 Provider を保持
type Provider struct {
	TracerProvider *sdktrace.TracerProvider
}

// NewProvider は OTEL Provider を初期化
func NewProvider(ctx context.Context, cfg Config) (*Provider, error) {
	// 1. リソースの定義 (全テレメトリ共通)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// 2. Exporter の作成 (標準出力に出力)
	// 本番環境では otlptracegrpc.New() を使用して OTLP Collector に送信
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	// 3. TracerProvider の作成
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithBatchTimeout(5*time.Second), // NOTE: 5秒ごとに出力
			sdktrace.WithMaxExportBatchSize(512),     // NOTE: または512件溜まったら出力
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // 開発環境用
	)

	// 4. グローバルに設定
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // traceparent, tracestate ヘッダー
		propagation.Baggage{},      // 追加のコンテキスト情報
	))

	return &Provider{TracerProvider: tp}, nil
}

// Shutdown は Provider を終了
func (p *Provider) Shutdown(ctx context.Context) error {
	return p.TracerProvider.Shutdown(ctx)
}

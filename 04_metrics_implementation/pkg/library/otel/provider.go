// pkg/otel/provider.go

package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
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
	MeterProvider  *sdkmetric.MeterProvider
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

	// 2. Trace Exporter の作成 (標準出力に出力)
	// 本番環境では otlptracegrpc.New() を使用して OTLP Collector に送信
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	// 3. TracerProvider の作成
	//
	// - WithBatcher: スパンを即時エクスポートせず、バッチに溜めてからまとめて送信する。
	//   SimpleSpanProcessor (即時送信) もあるが、本番ではバッチが推奨。
	//
	//   - WithBatchTimeout(5s): 最後のエクスポートから5秒経過したらバッチをフラッシュする。
	//     スパンが少量でも一定間隔でエクスポートされることを保証する。
	//
	//   - WithMaxExportBatchSize(512): バッチに512件溜まった時点で即座にエクスポートする。
	//     高トラフィック時にメモリ上にスパンが溜まりすぎるのを防ぐ。
	//     タイムアウトとバッチサイズの「どちらか先に到達した方」でエクスポートが発火する。
	//
	// - WithSampler: どのスパンを記録するかを制御する。
	//   AlwaysSample() は全リクエストのスパンを記録する (開発環境向け)。
	//   本番環境では TraceIDRatioBased(0.1) 等で10%だけ記録するなど、
	//   データ量とコストを抑えるサンプリング戦略を選択する。
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithBatchTimeout(5*time.Second), // NOTE: 5秒間隔でトレースを出力
			sdktrace.WithMaxExportBatchSize(512),     // NOTE: または512件溜まったら出力
		),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// 4. Metric Exporter の作成 (標準出力に出力)
	metricExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	// 5. MeterProvider の作成
	//
	// TracerProvider との違い:
	//
	// - BatchSize に相当する設定がない:
	//   トレースはスパン1件1件が独立したデータであり、バッチに溜めてからエクスポートする。
	//   メトリクスは PeriodicReader の収集タイミングで全メトリクスが集約済みの状態になる。
	//   例えば Counter.Add(1) を100回呼んでも、エクスポート時には Value:100 という
	//   1つのデータポイントになるため、「何件溜まったら送るか」という概念自体がない。
	//
	// - Sampler に相当する設定がない:
	//   トレースはリクエストごとにスパンが生成され、高トラフィック時にデータ量が膨大になるため
	//   サンプリング (一部だけ記録) が必要になる。
	//   メトリクスは上記の通り集約済みなので、リクエスト数が増えてもデータポイント数は増えない。
	//   Counter は常に1つの累計値、Histogram はバケットごとの固定数の値であるため、
	//   サンプリングの必要がなく、SDK にも API が存在しない。
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(metricExporter,
				sdkmetric.WithInterval(10*time.Second), // NOTE: 10秒間隔でメトリクスを収集・エクスポート
			),
		),
	)

	// 6. グローバルに設定
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // traceparent, tracestate ヘッダー
		propagation.Baggage{},      // 追加のコンテキスト情報
	))

	return &Provider{
		TracerProvider: tp,
		MeterProvider:  mp,
	}, nil
}

// Shutdown は Provider を終了
func (p *Provider) Shutdown(ctx context.Context) error {
	// TracerProvider → MeterProvider の順にシャットダウン
	if err := p.TracerProvider.Shutdown(ctx); err != nil {
		return err
	}
	return p.MeterProvider.Shutdown(ctx)
}

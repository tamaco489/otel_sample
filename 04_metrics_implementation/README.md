# 04_metrics_implementation

OpenTelemetry の Metrics API/SDK を使い、Counter・Histogram・Observable Gauge の3種類のメトリクスを計装するサンプル。

## 概要

`03_slog_otel_integration` ではトレース + slog 統合まで実装したが、メトリクスは未実装だった。
本サンプルでは `MeterProvider` を追加し、アプリケーションの定量的な状態を収集・エクスポートする。

## 03_slog_otel_integration との差分

### 1. `pkg/library/otel/provider.go` (変更)

`MeterProvider` を追加。`stdoutmetric` Exporter + `PeriodicReader` (10秒間隔) で構成。

```go
type Provider struct {
    TracerProvider *sdktrace.TracerProvider
    MeterProvider  *sdkmetric.MeterProvider  // ★ 追加
}
```

### 2. `internal/usecase/metrics.go` (新規)

パッケージレベルで `meter` とメトリクス計器を定義。

### 3. `internal/usecase/get_by_id.go` (変更)

記事取得成功時に `article.views.total` Counter をインクリメント。

### 4. `internal/usecase/create.go` (変更)

記事作成の処理時間を `article.create.duration` Histogram に記録。成功・バリデーションエラー・DBエラーの全パスで `status` 属性付きで記録。

### 5. `internal/repository/repository.go` / `create.go` (変更)

`publishedCount` (`atomic.Int64`) を導入し、記事作成成功時にインクリメント。Observable Gauge のコールバックで参照される。

## メトリクスの種類

| メトリクス名              | 種類             | 説明                     | 記録タイミング                |
| ------------------------- | ---------------- | ------------------------ | ----------------------------- |
| `article.views.total`     | Counter          | 記事閲覧数の累計         | GET /articles/{id} 成功時     |
| `article.create.duration` | Histogram        | 記事作成の処理時間（秒） | POST /articles 完了時         |
| `article.active.count`    | Observable Gauge | 公開中の記事数           | PeriodicReader のコールバック |

### Counter (累積カウンター)

単調増加する値。リセットされない。閲覧数・リクエスト数・エラー数などに使用。

```go
articleViewsCounter.Add(ctx, 1, metric.WithAttributes(...))
```

### Histogram (ヒストグラム)

値の分布を記録する。バケット境界を指定して分布を集計。レイテンシ・レスポンスサイズなどに使用。

```go
articleCreateDuration.Record(ctx, duration, metric.WithAttributes(...))
```

### Observable Gauge (観測可能ゲージ)

コールバックで定期的に観測される瞬時値。CPU使用率・メモリ使用量・キュー長などに使用。

```go
meter.Int64ObservableGauge("article.active.count",
    metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
        o.Observe(repository.GetPublishedCount())
        return nil
    }),
)
```

## MeterProvider 初期化フロー

```
Resource (サービスメタデータ: TracerProvider と共通)
    ↓
stdoutmetric.Exporter (メトリクスを JSON で標準出力)
    ↓
PeriodicReader (10秒間隔で収集・エクスポート)
    ↓
MeterProvider
    ↓
otel.SetMeterProvider() でグローバル登録
```

## 実行方法

```bash
# サーバー起動
go run ./cmd/main.go

# リクエスト送信
curl http://localhost:8080/articles/article-123
curl -X POST http://localhost:8080/articles \
  -H "Content-Type: application/json" \
  -d '{"Title":"OpenTelemetry Metrics入門","Content":"Counter, Histogram, Observable Gaugeの使い方"}'

# 10秒後に stdout にメトリクス JSON が出力される
# - article.views.total の Counter 値
# - article.create.duration の Histogram 値
# - article.active.count の Gauge 値
```

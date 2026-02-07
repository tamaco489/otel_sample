# 03_slog_otel_integration

`slog` と OpenTelemetry を統合し、構造化ログに `trace_id` / `span_id` を自動注入するサンプル。

## 概要

`02_instrumentation` では `slog.InfoContext(ctx, ...)` を使用しているが、ログ出力に `trace_id` / `span_id` が含まれない。
本サンプルでは、カスタム `slog.Handler`（`OTELHandler`）を実装し、OpenTelemetry の `SpanContext` からトレース情報をログに透過的に注入する。

## OTELHandler の仕組み

```
slog.InfoContext(ctx, "message")
    ↓
OTELHandler.Handle(ctx, record)
    ↓ SpanContext から trace_id / span_id を取得
    ↓ record.AddAttrs(trace_id, span_id)
    ↓
JSONHandler.Handle(ctx, record)  ← 元のハンドラに委譲
    ↓
{"time":"...","level":"INFO","msg":"message","trace_id":"abc...","span_id":"def..."}
```

## 実行方法

```bash
# サーバー起動
go run ./cmd/main.go

# リクエスト送信
curl http://localhost:8080/articles/article-123

# 記事作成
curl -X POST http://localhost:8080/articles \
  -H "Content-Type: application/json" \
  -d '{"Title":"OpenTelemetry入門","Content":"slogとOTELの統合について"}'
```

## 期待出力

ログ出力に `trace_id` と `span_id` が含まれる:

```json
{"time":"...","level":"INFO","msg":"article retrieved","id":"article-123","title":"サンプル記事","status":"published","trace_id":"abc123...","span_id":"def456..."}
```

同一リクエスト内の全ログで `trace_id` が一致し、リクエストの追跡が可能になる。

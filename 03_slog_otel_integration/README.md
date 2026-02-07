# 03_slog_otel_integration

`slog` と OpenTelemetry を統合し、構造化ログに `trace_id` / `span_id` を自動注入するサンプル。

## 概要

`02_instrumentation` では `slog.InfoContext(ctx, ...)` を使用しているが、ログ出力に `trace_id` / `span_id` が含まれない。
本サンプルでは、カスタム `slog.Handler`（`OTELHandler`）を実装し、OpenTelemetry の `SpanContext` からトレース情報をログに透過的に注入する。

## 02_instrumentation との差分

変更したのは **2ファイルのみ**。handler/usecase/repository のコードは一切変更していない。

### 1. `pkg/library/otel/handler.go` (新規)

`slog.Handler` インターフェースをラップするカスタムハンドラ。

```go
type OTELHandler struct {
    slog.Handler  // slog.Handler インターフェースの埋め込み
}
```

`OTELHandler` は `slog.Handler` を埋め込みで保持している。
埋め込みにより `slog.Handler` の4つのメソッド (`Enabled`, `Handle`, `WithAttrs`, `WithGroup`) が `OTELHandler` に昇格（プロモート）される。
そのうち `Handle`, `WithAttrs`, `WithGroup` を明示的に再定義（オーバーライド）し、独自の処理を追加している。

### 2. `cmd/main.go` (変更)

slog の初期化部分のみ変更。

```go
// JSONHandler を作成 (ログを JSON 形式で stdout に書き出すハンドラ)
jsonHandler := slog.NewJSONHandler(os.Stdout, nil)

// slog.Handler を内包した OTELHandler 構造体を生成
otelHandler := otel.NewOTELHandler(jsonHandler)

// Logger に登録
// 以降のビジネスロジックで slog.InfoContext などが実行された場合 OTELHandler.Handle が実行される
slog.SetDefault(slog.New(otelHandler))
```

## slog の構成要素

```
slog.Logger          ... ログ呼び出しの窓口 (InfoContext, ErrorContext 等)
    ↓ Record を渡す
slog.Handler         ... Record をどう処理するかを決める (インターフェース)
    ↓ 実装例:
    ├── JSONHandler    ... JSON 形式で io.Writer に書き出す
    ├── TextHandler    ... テキスト形式で io.Writer に書き出す
    └── OTELHandler    ... trace_id/span_id を追加して別の Handler に委譲する
```

`Handler` は Logger の「バックエンド」であり、ログをどこにどう出すかの戦略を担う。

## OTELHandler の処理フロー

ビジネスロジック中で `slog.InfoContext(ctx, ...)` が呼ばれると、以下の流れで処理される。

```
slog.InfoContext(ctx, "article retrieved", ...)
  ↓
Logger.InfoContext()
  ↓ Record を生成
OTELHandler.Handle(ctx, record)
  ↓ trace.SpanContextFromContext(ctx) で SpanContext を取得
  ↓ spanCtx.IsValid() が true なら trace_id / span_id を record に追加
  ↓
JSONHandler.Handle(ctx, record)  ← 内部ハンドラに委譲
  ↓
stdout に JSON 出力:
{"time":"...","level":"INFO","msg":"article retrieved","id":"...","trace_id":"...","span_id":"..."}
```

ポイント:

- `ctx` には otelhttp ミドルウェアが設定した SpanContext が含まれている
- `OTELHandler` はその `ctx` から `trace_id` / `span_id` を取り出してログレコードに注入する
- ビジネスロジック側は `slog.InfoContext(ctx, ...)` で `ctx` を渡すだけでよく、トレース情報を意識する必要がない

## 埋め込みによるメソッド解決

`OTELHandler` の `Handle` メソッドを削除（コメントアウト）してもコンパイルは通る。
これは Go の埋め込みにより、埋め込んだ `slog.Handler` (= `JSONHandler`) の `Handle` が自動的に使われるため。
ただしその場合 `trace_id` / `span_id` の注入が行われず、`02_instrumentation` と同じ出力になる。

| メソッド    | 明示的に定義? | 実際に呼ばれるもの                             |
| ----------- | ------------- | ---------------------------------------------- |
| `Enabled`   | なし          | 埋め込みの `JSONHandler.Enabled`               |
| `Handle`    | あり          | `OTELHandler.Handle` (trace_id/span_id を注入) |
| `WithAttrs` | あり          | `OTELHandler.WithAttrs` (再ラップして返す)     |
| `WithGroup` | あり          | `OTELHandler.WithGroup` (再ラップして返す)     |

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

## 出力の比較

**02_instrumentation (trace_id/span_id なし):**

```json
{
  "time": "...",
  "level": "INFO",
  "msg": "article retrieved",
  "id": "article-123",
  "title": "サンプル記事",
  "status": "published"
}
```

**03_slog_otel_integration (trace_id/span_id あり):**

```json
{
  "time": "...",
  "level": "INFO",
  "msg": "article retrieved",
  "id": "article-123",
  "title": "サンプル記事",
  "status": "published",
  "trace_id": "c1711038cb53c863f03d5b1826bc3b20",
  "span_id": "6c9d7c4e21b41527"
}
```

同一リクエスト内の全ログで `trace_id` が一致し、リクエスト単位でのログ追跡が可能になる。
また `trace_id` はトレーススパンの `TraceID` と一致するため、ログとトレースの紐付けができる。

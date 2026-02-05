# 02_instrumentation

OpenTelemetry 手動計装のサンプル

## 実行

```bash
go run ./cmd/main.go
```

## 構成

```
cmd/main.go           エントリポイント (OTEL初期化、DI)
internal/
  config/             アプリケーション設定
  controller/         サーバー実行制御
  di/                 依存関係の組み立て
  handler/            リクエスト処理、ログ出力
  usecase/            ビジネスロジック、スパン計装
  repository/         データアクセス、DB計装
pkg/
  errors/             エラー定義
  library/otel/       OTEL Provider
```

## 依存関係

```
main → di → controller → handler → usecase → repository
```

## 計装ポイント

| レイヤー   | SpanKind | 内容                           |
| ---------- | -------- | ------------------------------ |
| HTTP       | Server   | otelhttp 自動計装              |
| usecase    | Internal | ビジネスロジック、イベント記録 |
| repository | Client   | DB操作 (SELECT/INSERT)         |

## トレース出力の仕組み

```
リクエスト処理
    ↓
span.End() → BatchSpanProcessor のキューに追加
    ↓
 (5秒待機 or 512件蓄積)
    ↓
stdouttrace Exporter → JSON出力
```

| 方式     | メリット         | デメリット    |
| -------- | ---------------- | ------------- |
| 即時出力 | リアルタイム     | I/O負荷が高い |
| バッチ   | 効率的、本番向け | 遅延がある    |

本番環境では OTLP Collector に送信するため、バッチ処理が効率的。

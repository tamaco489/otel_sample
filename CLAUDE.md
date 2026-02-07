# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

Go で OpenTelemetry の Traces / Metrics を段階的に実装する学習用サンプルプロジェクト。
各サンプルは独立した Go モジュールとして構成されている。

## ビルド・実行コマンド

```bash
# 各サンプルディレクトリに移動してから実行する
cd <サンプル名> && go run ./cmd/main.go

# 依存関係のダウンロード
cd <サンプル名> && go mod download

# ビルド・静的解析
cd <サンプル名> && go build ./cmd/main.go && go vet ./...
```

## プロジェクト構成

段階的に機能を追加する構成:

- **01_start_otel**: 基本的な OTEL セットアップ (Exporter → Resource → TracerProvider → Tracer → Spans)
- **02_instrumentation**: レイヤードアーキテクチャと手動スパン計装を含む HTTP API
- **03_slog_otel_integration**: slog + OTEL 統合。ログに trace_id / span_id を自動付与
- **04_metrics_implementation**: MeterProvider + Counter / Histogram / Observable Gauge によるメトリクス計装

## アーキテクチャ (02〜04 共通)

```
main → di → controller → handler → usecase → repository
```

| レイヤー          | SpanKind | 計装方式                   |
| ----------------- | -------- | -------------------------- |
| HTTP (controller) | Server   | otelhttp 自動計装          |
| usecase           | Internal | 手動スパン、イベント、属性 |
| repository        | Client   | DB 操作用の手動スパン      |

## 主要ディレクトリ構成 (02〜04 共通)

```
<サンプル名>/
├── cmd/main.go                     # エントリポイント
├── internal/
│   ├── config/                     # アプリケーション設定
│   ├── controller/                 # HTTP サーバー (otelhttp)
│   ├── di/                         # 依存注入
│   ├── entity/                     # エンティティ定義 (04〜)
│   ├── handler/                    # リクエストハンドラ
│   ├── repository/                 # データアクセス層
│   └── usecase/                    # ビジネスロジック・メトリクス計装
├── pkg/library/otel/               # OTEL Provider 初期化
└── sample/                         # 検証用リクエスト・出力例
```

## テレメトリ出力

- **Traces**: 開発環境では stdouttrace で JSON 出力 (BatchSpanProcessor)
- **Metrics**: 開発環境では stdoutmetric で JSON 出力 (PeriodicReader)
- 本番環境では OTLP Collector へバイナリ (protobuf) 送信を想定

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

Go で分散トレーシングを実装する OpenTelemetry 学習用サンプルプロジェクト。

## ビルド・実行コマンド

```bash
# 01_start_otel - 基本的なトレーシング例
go run ./01_start_otel/cmd/main.go

# 02_instrumentation - 手動計装を含む完全な Web アプリケーション
cd 02_instrumentation && go run ./cmd/main.go

# 依存関係のダウンロード（各モジュール個別）
cd 01_start_otel && go mod download
cd 02_instrumentation && go mod download
```

## プロジェクト構成

2つの段階的なサンプル:

- **01_start_otel**: 基本的な OTEL セットアップ（Exporter → Resource → TracerProvider → Tracer → Spans）
- **02_instrumentation**: レイヤードアーキテクチャと手動スパン計装を含む完全な HTTP API

## アーキテクチャ (02_instrumentation)

```
main → di → controller → handler → usecase → repository
```

| レイヤー          | SpanKind | 計装方式                   |
| ----------------- | -------- | -------------------------- |
| HTTP (controller) | Server   | otelhttp 自動計装          |
| usecase           | Internal | 手動スパン、イベント、属性 |
| repository        | Client   | DB 操作用の手動スパン      |

API エンドポイント:

- `GET /articles/{id}` - 記事取得
- `POST /articles` - 記事作成

## OTEL セットアップパターン

```
Resource (サービスメタデータ)
    ↓
Exporter (開発: stdouttrace、本番: OTLP)
    ↓
TracerProvider (BatchSpanProcessor: 5秒タイムアウト、最大512件)
    ↓
グローバル登録 (otel.SetTracerProvider)
```

## 主要ファイル

- `02_instrumentation/pkg/library/otel/provider.go` - OTEL プロバイダー初期化
- `02_instrumentation/internal/usecase/*.go` - 手動スパン計装の例
- `02_instrumentation/internal/controller/server.go` - otelhttp ミドルウェア設定

## トレース出力

BatchSpanProcessor により出力は最大5秒遅延する。バッチフラッシュまたはシャットダウン時に JSON 形式で stdout に出力される。

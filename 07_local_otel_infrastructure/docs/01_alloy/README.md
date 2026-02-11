# Grafana Alloy

## Alloy とは

Grafana Alloy は、アプリケーションからトレース・メトリクス・ログを受け取り、適切なバックエンドへ転送する統合テレメトリコレクター。OpenTelemetry Collector と Promtail を 1 つのプロセスで置き換える役割を持つ。

## なぜ Alloy を採用したか

- 通常、OTLP データの収集には OpenTelemetry Collector、Docker ログの収集には Promtail と **2 つの別プロセス** が必要になる。Alloy はこれらを **1 コンテナで統合** できるため、構成がシンプルになる
- Grafana スタック (Tempo / Loki / Prometheus) との親和性が高く、設定の記述量も少ない
- Alloy 独自の設定言語 (HCL ライク) により、パイプラインの流れを宣言的に記述できる

## 設定ファイル

- パス: `_docker/alloy/config.alloy`

## 全体構成

この設定では 2 つの独立したパイプラインが定義されている。

```text
パイプライン 1: トレース & メトリクス (OTLP)

  アプリ (Go)
    │  OTLP プロトコルで送信
    ▼
  [receiver.otlp]     ← gRPC(:4317) / HTTP(:4318) で受信
    │
    ▼
  [processor.batch]    ← 5秒 or 1024件でバッファリング
    │
    ├── traces  → [exporter.otlp "tempo"]        → Tempo
    └── metrics → [exporter.prometheus "default"] → Prometheus


パイプライン 2: コンテナログ

  Docker コンテナ (stdout/stderr)
    │  docker.sock 経由
    ▼
  [discovery.docker]      ← 5秒ごとにコンテナを自動検出
    │
    ▼
  [discovery.relabel]     ← メタデータをラベルに変換、Alloy 自身を除外
    │
    ▼
  [loki.source.docker]    ← コンテナログを読み取り
    │
    ▼
  [loki.process]          ← JSON パースし trace_id / span_id / level を抽出
    │
    ▼
  [loki.write]            → Loki
```

## 各ブロックの解説

### パイプライン 1: トレース & メトリクス

#### `otelcol.receiver.otlp "default"`

アプリからの OTLP データを 2 つのポートで待ち受ける。

| プロトコル | ポート | 用途                               |
| ---------- | ------ | ---------------------------------- |
| gRPC       | 4317   | Go / Java SDK が主に使うプロトコル |
| HTTP       | 4318   | gRPC が使えない環境向けの代替手段  |

受信データは batch プロセッサへ渡される。

**なぜ gRPC と HTTP の両方を開けるのか:** Go SDK はデフォルトで gRPC を使うが、ブラウザや一部の環境では gRPC が使えないため HTTP も用意しておく。両方開けるコストはほぼゼロなので、将来の拡張性のために両対応にしている。

#### `otelcol.processor.batch "default"`

テレメトリデータをバッファリングし、まとめて送信することでスループットを向上させる。

| パラメータ        | 値     | 説明                           |
| ----------------- | ------ | ------------------------------ |
| `timeout`         | `5s`   | 5 秒ごとにバッファをフラッシュ |
| `send_batch_size` | `1024` | 1024 件溜まったらフラッシュ    |

トレースは Tempo、メトリクスは Prometheus へルーティングされる。

**なぜバッチ処理を挟むのか:** 1 件ずつ即座に転送すると、ネットワークオーバーヘッドが大きくなる。バッチにまとめることで転送回数を減らし、バックエンドへの負荷を軽減する。5 秒 / 1024 件はローカル開発で十分なレイテンシとスループットのバランス。

#### `otelcol.exporter.otlp "tempo"`

トレースデータを gRPC (`tempo:4317`) で Tempo へ転送する。
`insecure = true` はローカル Docker ネットワーク内のため TLS を無効化する設定。

**なぜ `insecure = true` なのか:** Docker Compose の内部ネットワークでは通信が外部に出ないため、TLS による暗号化は不要。設定を簡潔に保つために無効化している。本番環境では TLS を有効にすること。

#### `otelcol.exporter.prometheus "default"` / `prometheus.remote_write "default"`

OTLP メトリクスを Prometheus 形式に変換し、Remote Write API (`http://prometheus:9090/api/v1/write`) 経由で Prometheus に書き込む。

**なぜ Remote Write (プッシュ) を使うのか:** Prometheus の標準的なメトリクス収集はプル型 (Prometheus がアプリをスクレイプする) だが、OTLP で送られるメトリクスはプッシュ型。Alloy が OTLP → Prometheus 形式の変換を担い、Remote Write で書き込むことで、アプリ側は OTLP のみに統一できる。

### パイプライン 2: コンテナログ

#### `discovery.docker "containers"`

Docker ソケットに接続し、5 秒間隔で稼働中のコンテナを自動検出する。

**なぜ自動検出するのか:** コンテナ名やポートを設定ファイルにハードコードすると、コンテナの追加・削除のたびに設定変更が必要になる。Docker ソケット経由の自動検出なら、新しいコンテナが起動すれば自動的にログ収集が始まる。

#### `discovery.relabel "docker_logs"`

コンテナのメタデータを加工する。

| ルール                                                                 | 目的                                              |
| ---------------------------------------------------------------------- | ------------------------------------------------- |
| `alloy` サービスを除外                                                 | Alloy 自身のログ収集を防止 (無限ループ回避)       |
| `__meta_docker_container_name` → `container`                           | コンテナ名を `container` ラベルにマッピング       |
| `__meta_docker_container_label_com_docker_compose_service` → `service` | Compose サービス名を `service` ラベルにマッピング |

**なぜ Alloy 自身を除外するのか:** Alloy がログを収集 → そのログ出力を Alloy が再度収集 → 再びログが出力される...という無限ループが発生するため。

**なぜラベルをマッピングするのか:** Docker のメタデータキー (`__meta_docker_container_name` 等) はそのままでは Loki のラベルとして使えない。`container` や `service` のような短い名前に変換することで、Grafana 上で `{service="article-server"}` のように直感的にフィルタリングできるようになる。

#### `loki.source.docker "default"`

Docker ソケット経由で検出されたコンテナからログを読み取り、処理ステージへ渡す。

#### `loki.process "docker_logs"`

ログ行を JSON としてパースし、フィールドを Loki ラベルに昇格させる。

| 抽出フィールド | 目的                                           |
| -------------- | ---------------------------------------------- |
| `trace_id`     | Grafana でトレースとログの相互参照を可能にする |
| `span_id`      | スパン単位のログ絞り込みを可能にする           |
| `level`        | ログレベルでのフィルタリングを可能にする       |

**なぜこの 3 フィールドを抽出するのか:**

- `trace_id` / `span_id`: Grafana の Tempo ↔ Loki 連携のキーとなるフィールド。これをラベルに昇格させることで、トレース画面から「このスパンに対応するログ」を即座に表示できる。03_slog_otel_integration で slog に trace_id / span_id を埋め込む実装をしているのは、ここで活用するため
- `level`: `{level="error"}` のようにクエリできるようにすることで、障害調査時にエラーログだけを素早く絞り込める

#### `loki.write "default"`

加工済みログを `http://loki:3100/loki/api/v1/push` へ送信する。

## 公開ポート

| ポート | プロトコル | 用途                              |
| ------ | ---------- | --------------------------------- |
| 4317   | gRPC       | OTLP 受信 (トレース & メトリクス) |
| 4318   | HTTP       | OTLP 受信 (トレース & メトリクス) |
| 12345  | HTTP       | Alloy 管理 UI                     |

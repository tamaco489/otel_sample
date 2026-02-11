# Grafana

## Grafana とは

Grafana は可視化・ダッシュボードプラットフォーム。複数のデータソース (Tempo, Prometheus, Loki) に接続し、トレース・メトリクス・ログを横断的に探索できる統合インターフェースを提供する。

## なぜ Grafana を採用したか

- **統合可視化:** トレース (Tempo)、メトリクス (Prometheus)、ログ (Loki) を **1 つの画面で横断的に** 確認できる。ツールごとに別の UI を開く必要がない
- **データソース間の相互参照:** トレースからログへ、ログからトレースへワンクリックで遷移できる仕組みが組み込まれている。障害調査の効率が大幅に向上する
- **プロビジョニング:** 設定ファイルでデータソースを宣言的に管理できるため、`docker-compose up` だけで設定済みの Grafana が立ち上がる

## 設定ファイル

- パス: `_docker/grafana/provisioning/datasources/datasources.yaml`

Grafana はプロビジョニングファイルを使い、起動時にデータソースを自動設定する。Web UI での手動設定が不要になる。

**なぜプロビジョニングを使うのか:** Web UI で手動設定すると、コンテナを再作成するたびに設定がリセットされる。プロビジョニングファイルならコードとして管理でき、誰がいつ `docker-compose up` しても同じ環境が再現される。

## 環境変数 (docker-compose)

```yaml
environment:
  - GF_AUTH_ANONYMOUS_ENABLED=true
  - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
  - GF_AUTH_DISABLE_LOGIN_FORM=true
```

| 変数                         | 値      | 説明                                 |
| ---------------------------- | ------- | ------------------------------------ |
| `GF_AUTH_ANONYMOUS_ENABLED`  | `true`  | ログインなしでアクセス可能にする     |
| `GF_AUTH_ANONYMOUS_ORG_ROLE` | `Admin` | 匿名ユーザーに Admin 権限を付与      |
| `GF_AUTH_DISABLE_LOGIN_FORM` | `true`  | ログインフォームを完全に非表示にする |

**なぜ認証を無効にするのか:** ローカル開発環境ではログイン操作が手間になるだけで、セキュリティ上の恩恵がない。Admin 権限にしているのは、データソースの設定変更やダッシュボードの作成を制限なく行えるようにするため。本番環境では必ず適切な認証を設定すること。

## データソース設定

### Tempo

```yaml
- name: Tempo
  type: tempo
  url: http://tempo:3200
  uid: tempo
  isDefault: true
```

デフォルトのデータソースとして設定。他のデータソースとの相互参照が構成されている。

| 機能              | 連携先                    | 説明                                                                                             |
| ----------------- | ------------------------- | ------------------------------------------------------------------------------------------------ |
| `tracesToLogs`    | Loki (`loki`)             | トレースのスパンをクリックすると対応するログへジャンプ。`trace_id` と `span_id` でフィルタリング |
| `tracesToMetrics` | Prometheus (`prometheus`) | トレースから関連するメトリクスへ遷移                                                             |
| `nodeGraph`       | —                         | サービス間の依存関係を可視化するノードグラフを有効化                                             |
| `serviceMap`      | Prometheus (`prometheus`) | メトリクスデータからサービスマップを生成                                                         |

**なぜ `isDefault: true` なのか:** このプロジェクトではトレースを起点とした調査フローが主なユースケースのため、Tempo をデフォルトにしている。Grafana の Explore 画面を開いたときに最初に Tempo が選択された状態になる。

**なぜ `tracesToLogs` で `filterByTraceID` と `filterBySpanID` を有効にするのか:** トレースの特定のスパンをクリックしたとき、そのスパンに対応するログだけに絞り込んで表示するため。03_slog_otel_integration で slog に `trace_id` / `span_id` を埋め込んでおり、Alloy の `loki.process` でラベルに昇格させているのは、この連携を実現するため。

### Prometheus

```yaml
- name: Prometheus
  type: prometheus
  url: http://prometheus:9090
  uid: prometheus
```

PromQL でメトリクスをクエリするための標準的な Prometheus データソース。

**なぜ `uid: prometheus` を明示するのか:** Tempo の `tracesToMetrics` や `serviceMap` から Prometheus を参照する際に `datasourceUid` を使う。uid を明示しないと Grafana が自動採番するため、プロビジョニングファイル間の参照が不安定になる。

### Loki

```yaml
- name: Loki
  type: loki
  url: http://loki:3100
  uid: loki
```

トレースとの紐付けのために derived field が設定されている。

```yaml
jsonData:
  derivedFields:
    - datasourceUid: tempo
      matcherRegex: "\"trace_id\":\"(\\w+)\""
      name: TraceID
      url: "$${__value.raw}"
      urlDisplayLabel: View Trace
```

| パラメータ        | 説明                                              |
| ----------------- | ------------------------------------------------- |
| `matcherRegex`    | JSON ログ行から `trace_id` の値を抽出する正規表現 |
| `datasourceUid`   | 抽出したトレース ID を Tempo データソースにリンク |
| `urlDisplayLabel` | マッチしたログ行の横に "View Trace" リンクを表示  |

**なぜ derived field を設定するのか:** Tempo 側の `tracesToLogs` はトレース → ログの方向のみ。逆方向 (ログ → トレース) を実現するには、Loki 側でログ行から `trace_id` を抽出し、Tempo へのリンクを生成する必要がある。これにより **双方向の遷移** が完成する。

## 相互参照の全体像

```text
Tempo ──tracesToLogs──→ Loki       (トレース → ログ)
Tempo ──tracesToMetrics→ Prometheus (トレース → メトリクス)
Loki  ──derivedFields──→ Tempo     (ログ → トレース)
```

3 つのデータソース間の双方向リンクにより、Grafana 上でトレース・メトリクス・ログをシームレスに行き来できる。

**なぜ双方向リンクが重要なのか:** 障害調査では「エラーログを発見 → そのリクエストのトレースを確認 → 遅延しているスパンを特定 → そのスパンのログを深掘り」のように、データソース間を頻繁に行き来する。リンクがなければ毎回 `trace_id` を手動でコピーして検索する必要があり、調査速度が大幅に落ちる。

## 公開ポート

| ポート | プロトコル | 用途           |
| ------ | ---------- | -------------- |
| 3000   | HTTP       | Grafana Web UI |

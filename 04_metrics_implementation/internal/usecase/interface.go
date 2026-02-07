package usecase

import (
	"context"

	"github.com/tamaco489/otel_sample/04_metrics_implementation/internal/entity"
	"github.com/tamaco489/otel_sample/04_metrics_implementation/internal/repository"
	"go.opentelemetry.io/otel/metric"
)

// ArticleUsecase は記事ユースケースのインターフェース
type ArticleUsecase interface {
	GetByID(ctx context.Context, id string) (*entity.Article, error)
	Create(ctx context.Context, input *CreateArticleInput) (*entity.Article, error)
}

// articleUsecase は ArticleUsecase の実装
type articleUsecase struct {
	repo repository.ArticleRepository
}

// NewArticleUsecase は ArticleUsecase を生成し、メトリクス計器を初期化する。
//
// meter.Int64Counter() 等は (instrument, error) の2値を返すため、パッケージレベルの var 宣言では直接受け取れない。
// otel.Tracer() / otel.Meter() はエラーを返さない1値なので var で宣言できる。
//
// なお otel.Meter() は SetMeterProvider() より前に呼んでも安全。
// 取得した Meter はグローバル委譲パターンにより、後から設定された実際の MeterProvider に自動的に委譲される。(otel.Tracer() と同じ仕組み)
func NewArticleUsecase(repo repository.ArticleRepository) ArticleUsecase {
	var err error

	// Counter: 記事閲覧数の累計
	//
	// Counter は単調増加する値を記録するメトリクス。Add() で加算のみ可能で、減算やリセットはできない。
	// 閲覧数・リクエスト数・エラー数など「累計で何回起きたか」を数える用途に使用する。
	//
	// Int64Counter: 整数値の Counter。閲覧1回につき Add(ctx, 1) で +1 する。
	// (Float64Counter もあるが、回数のカウントには整数が適切)
	//
	// エクスポート時には CumulativeTemporality の場合「起動時からの累計値」が送信される。
	// 例: 10秒間に5回閲覧 → Value: 5、さらに3回閲覧 → Value: 8 (差分ではなく累計)
	// Prometheus 等のバックエンドでは rate() 関数で「毎秒の増加率」に変換してグラフ化する。
	articleViewsCounter, err = meter.Int64Counter(
		"article.views.total",
		metric.WithDescription("記事閲覧数の累計"),
	)
	if err != nil {
		panic(err)
	}

	// Histogram: 記事作成の処理時間 (秒)
	//
	// Histogram は Record() で渡された値を、事前に定義したバケット境界に振り分けて分布を記録する。
	// Counter が「何回起きたか」を数えるのに対し、Histogram は「どのくらいの時間がかかったか」の分布を可視化する。
	//
	// WithExplicitBucketBoundaries で指定した境界値により、以下の6つのバケットが生成される:
	//   [0, 0.1) / [0.1, 0.5) / [0.5, 1) / [1, 2) / [2, 5) / [5, +Inf)
	// 例: Record(0.3) → [0.1, 0.5) バケットのカウントが +1
	//     Record(2.1) → [2, 5) バケットのカウントが +1
	//
	// エクスポート時には各バケットのカウント・合計値・最小値・最大値がまとめて送信される。
	// これにより「リクエストの95%が0.5秒以内に完了しているか」等のパーセンタイル分析が可能になる。
	//
	// バケット境界の設計指針:
	//   - SLO/SLA の閾値を含める (例: 目標レイテンシが1秒なら 1 を含める)
	//   - 境界を細かくしすぎるとデータポイント数が増えるため、5〜10個程度が目安
	//   - 未指定の場合は SDK デフォルト境界が使われるが、アプリ特性に合わせて明示指定を推奨
	articleCreateDuration, err = meter.Float64Histogram(
		"article.create.duration",
		metric.WithDescription("記事作成の処理時間 (秒)"),
		metric.WithExplicitBucketBoundaries(0.1, 0.5, 1, 2, 5), // NOTE: 6つのバケットに分布を記録
	)
	if err != nil {
		panic(err)
	}

	// Observable Gauge: 公開中の記事数
	//
	// Observable Gauge は「現在の瞬時値」を定期的に観測するメトリクス。
	// Counter (累積) や Histogram (分布) と異なり、値は増減する (例: 公開記事数、CPU使用率、キュー長)。
	//
	// Counter/Histogram は「同期的」: ビジネスロジック中で Add()/Record() を明示的に呼ぶ。
	// Observable Gauge は「非同期的」: コールバックを1回登録するだけで、あとは PeriodicReader が収集タイミング (10秒間隔) ごとにコールバックを呼び出す。
	// そのため、リクエストの有無に関係なく常に最新の値が観測される。
	//
	// コールバック内では ArticleRepository インターフェース経由で値を取得する。
	// repo はクロージャでキャプチャされ、PeriodicReader のコールバック呼び出し時に参照される。
	_, err = meter.Int64ObservableGauge(
		"article.active.count",
		metric.WithDescription("公開中の記事数"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(repo.GetPublishedCount(ctx)) // NOTE: インターフェース経由で現在の公開記事数を観測
			return nil
		}),
	)
	if err != nil {
		panic(err)
	}

	return &articleUsecase{repo: repo}
}

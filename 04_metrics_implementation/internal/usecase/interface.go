package usecase

import (
	"context"

	"github.com/tamaco489/otel_sample/04_metrics_implementation/internal/repository"
	"go.opentelemetry.io/otel/metric"
)

// ArticleUsecase は記事ユースケースのインターフェース
type ArticleUsecase interface {
	GetByID(ctx context.Context, id string) (*repository.Article, error)
	Create(ctx context.Context, input *CreateArticleInput) (*repository.Article, error)
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
	articleViewsCounter, err = meter.Int64Counter(
		"article.views.total",
		metric.WithDescription("記事閲覧数の累計"),
	)
	if err != nil {
		panic(err)
	}

	// Histogram: 記事作成の処理時間（秒）
	articleCreateDuration, err = meter.Float64Histogram(
		"article.create.duration",
		metric.WithDescription("記事作成の処理時間（秒）"),
		metric.WithExplicitBucketBoundaries(0.1, 0.5, 1, 2, 5),
	)
	if err != nil {
		panic(err)
	}

	// Observable Gauge: 公開中の記事数
	// Counter/Histogram はロジック中で Add()/Record() を呼んで記録するが、
	// Observable Gauge はコールバックを1回登録するだけで、あとは PeriodicReader が定期的に呼び出す。
	_, err = meter.Int64ObservableGauge(
		"article.active.count",
		metric.WithDescription("公開中の記事数"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			o.Observe(repo.GetPublishedCount(ctx))
			return nil
		}),
	)
	if err != nil {
		panic(err)
	}

	return &articleUsecase{repo: repo}
}

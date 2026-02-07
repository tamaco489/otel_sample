package usecase

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/tamaco489/otel_sample/04_metrics_implementation/internal/entity"
	apperrors "github.com/tamaco489/otel_sample/04_metrics_implementation/pkg/errors"
)

// GetByID は記事をIDで取得する (手動計装の例)
func (u *articleUsecase) GetByID(ctx context.Context, id string) (*entity.Article, error) {
	// スパンの開始
	// SpanKindInternal: 内部処理を表す (デフォルト)
	ctx, span := tracer.Start(ctx, "ArticleUsecase.GetByID",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	// ビジネス固有の属性を追加
	span.SetAttributes(
		attribute.String("article.id", id),
	)

	// リポジトリ呼び出し
	article, err := u.repo.FindByID(ctx, id)
	if err != nil {
		// エラーを記録
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if article == nil {
		span.SetStatus(codes.Error, "article not found")
		return nil, apperrors.ErrNotFound
	}

	// 成功時の追加情報
	span.SetAttributes(
		attribute.String("article.title", article.Title),
		attribute.String("article.status", article.Status),
	)

	// メトリクス: 記事閲覧数をインクリメント
	//
	// NOTE: カーディナリティ (属性の組み合わせ数) に注意。
	// メトリクスの属性は「値の種類が有限」であることが前提。属性の組み合わせごとに独立した時系列 (データポイント) が生成されるため、ユニークな値が無制限に増える。
	// 属性 (例: article_id) を使うと時系列が爆発し、バックエンドのストレージとクエリ性能に深刻な影響を与える (カーディナリティ爆発)。
	//
	// 本サンプルでは学習目的で article_id を使用しているが、本番では以下のように低カーディナリティ属性 (status, category 等) のみを使い、
	// article_id のような高カーディナリティ属性はトレース (span attribute) で記録する。
	articleViewsCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("article_id", article.ID),
			attribute.String("category_id", "general"),
		),
	)
	// 本番例:
	//   articleViewsCounter.Add(ctx, 1,
	//       metric.WithAttributes(
	//           attribute.String("status", "published"),  // 数パターンに限定
	//           attribute.String("category", "tech"),      // 数十パターン程度
	//       ),
	//   )

	return article, nil
}

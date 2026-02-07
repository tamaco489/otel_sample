package repository

import (
	"context"
	"math/rand"

	"github.com/tamaco489/otel_sample/04_metrics_implementation/internal/entity"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// FindByID は記事をIDで取得する
func (r *articleRepository) FindByID(ctx context.Context, id string) (*entity.Article, error) {
	// 子スパンを作成
	// SpanKindClient: 外部サービス (DB等) への呼び出しを表す
	_, span := tracer.Start(ctx, "ArticleRepository.FindByID",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.String("article.id", id),
		),
	)
	defer span.End()

	// 模擬: 30%の確率で見つからない
	if rand.Float64() < 0.3 {
		return nil, nil
	}

	return &entity.Article{
		ID:     id,
		Title:  "サンプル記事",
		Status: "published",
	}, nil
}

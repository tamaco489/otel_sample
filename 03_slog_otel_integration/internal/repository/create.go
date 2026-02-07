package repository

import (
	"context"
	"errors"
	"math/rand"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ErrDBConnection はDB接続エラー
var ErrDBConnection = errors.New("database connection error")

// Create は記事を作成する
func (r *articleRepository) Create(ctx context.Context, article *Article) (*Article, error) {
	_, span := tracer.Start(ctx, "ArticleRepository.Create",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "INSERT"),
		),
	)
	defer span.End()

	// 模擬: 20%の確率でDB接続エラー
	if rand.Float64() < 0.2 {
		return nil, ErrDBConnection
	}

	// 模擬: IDを付与して返す
	article.ID = "new-article-456"
	article.Status = "published"

	span.SetAttributes(attribute.String("article.id", article.ID))

	return article, nil
}

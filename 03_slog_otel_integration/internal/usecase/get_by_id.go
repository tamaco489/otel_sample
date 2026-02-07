package usecase

import (
	"context"

	"github.com/tamaco489/otel_sample/03_slog_otel_integration/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	apperrors "github.com/tamaco489/otel_sample/03_slog_otel_integration/pkg/errors"
)

// GetByID は記事をIDで取得する (手動計装の例)
func (u *articleUsecase) GetByID(ctx context.Context, id string) (*repository.Article, error) {
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

	return article, nil
}

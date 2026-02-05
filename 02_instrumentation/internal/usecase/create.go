package usecase

import (
	"context"
	"fmt"

	"github.com/tamaco489/otel_sample/02_instrumentation/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	apperrors "github.com/tamaco489/otel_sample/02_instrumentation/pkg/errors"
)

// CreateArticleInput は記事作成の入力
type CreateArticleInput struct {
	Title   string
	Content string
}

// Validate は入力を検証
func (i *CreateArticleInput) Validate() error {
	if i.Title == "" {
		return fmt.Errorf("%w: title is required", apperrors.ErrValidation)
	}
	return nil
}

// Create は記事を作成する (イベント記録の例)
func (u *articleUsecase) Create(ctx context.Context, input *CreateArticleInput) (*repository.Article, error) {
	ctx, span := tracer.Start(ctx, "ArticleUsecase.Create")
	defer span.End()

	// イベント: バリデーション開始
	span.AddEvent("validation_started")

	if err := input.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return nil, err
	}

	// イベント: バリデーション完了
	span.AddEvent("validation_completed")

	// イベント: 作成開始
	span.AddEvent("create_started")

	// リポジトリで作成
	article := &repository.Article{
		Title: input.Title,
	}
	created, err := u.repo.Create(ctx, article)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// イベント: 作成完了 (属性付き)
	span.AddEvent("create_completed", trace.WithAttributes(
		attribute.String("article.id", created.ID),
	))

	return created, nil
}

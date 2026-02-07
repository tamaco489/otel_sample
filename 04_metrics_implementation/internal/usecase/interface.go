package usecase

import (
	"context"

	"github.com/tamaco489/otel_sample/04_metrics_implementation/internal/repository"
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

// NewArticleUsecase は ArticleUsecase を生成
func NewArticleUsecase(repo repository.ArticleRepository) ArticleUsecase {
	return &articleUsecase{repo: repo}
}

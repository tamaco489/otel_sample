package repository

import (
	"context"

	"github.com/tamaco489/otel_sample/04_metrics_implementation/internal/entity"
)

// ArticleRepository は記事リポジトリのインターフェース
type ArticleRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Article, error)
	Create(ctx context.Context, article *entity.Article) (*entity.Article, error)

	// GetPublishedCount は公開中の記事数を返す (Observable Gauge のコールバックから呼ばれる)
	GetPublishedCount(ctx context.Context) int64
}

// articleRepository は ArticleRepository の実装
type articleRepository struct{}

// NewArticleRepository は ArticleRepository を生成
func NewArticleRepository() ArticleRepository {
	return &articleRepository{}
}

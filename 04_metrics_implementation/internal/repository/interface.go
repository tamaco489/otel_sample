package repository

import "context"

// ArticleRepository は記事リポジトリのインターフェース
type ArticleRepository interface {
	FindByID(ctx context.Context, id string) (*Article, error)
	Create(ctx context.Context, article *Article) (*Article, error)

	// GetPublishedCount は公開中の記事数を返す (Observable Gauge のコールバックから呼ばれる)
	GetPublishedCount(ctx context.Context) int64
}

// articleRepository は ArticleRepository の実装
type articleRepository struct{}

// NewArticleRepository は ArticleRepository を生成
func NewArticleRepository() ArticleRepository {
	return &articleRepository{}
}

package repository

import "context"

// ArticleRepository は記事リポジトリのインターフェース
type ArticleRepository interface {
	FindByID(ctx context.Context, id string) (*Article, error)
	Create(ctx context.Context, article *Article) (*Article, error)
}

// articleRepository は ArticleRepository の実装
type articleRepository struct{}

// NewArticleRepository は ArticleRepository を生成
func NewArticleRepository() ArticleRepository {
	return &articleRepository{}
}

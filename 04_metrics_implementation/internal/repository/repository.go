package repository

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("repository/article")

// publishedCount は公開中の記事数 (本番では DB の COUNT クエリで取得する値をインメモリで代用)
var publishedCount atomic.Int64

// GetPublishedCount は公開中の記事数を返す (本番では DB クエリに置き換える)
func (r *articleRepository) GetPublishedCount(_ context.Context) int64 {
	return publishedCount.Load()
}

// Article は記事エンティティ
type Article struct {
	ID     string
	Title  string
	Status string
}

package handler

import (
	"github.com/tamaco489/otel_sample/02_instrumentation/internal/usecase"
)

// ArticleHandler は記事ハンドラ
type ArticleHandler struct {
	usecase usecase.ArticleUsecase
}

// NewArticleHandler は ArticleHandler を生成
func NewArticleHandler(uc usecase.ArticleUsecase) *ArticleHandler {
	return &ArticleHandler{usecase: uc}
}

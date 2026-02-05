package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/tamaco489/otel_sample/02_instrumentation/internal/usecase"
)

// CreateArticle は記事作成のHTTPハンドラ
// POST /articles
func (h *ArticleHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input usecase.CreateArticleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		slog.ErrorContext(ctx, "failed to decode request body",
			slog.String("error", err.Error()),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	article, err := h.usecase.Create(ctx, &input)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create article",
			slog.String("error", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "article created",
		slog.String("id", article.ID),
		slog.String("title", article.Title),
		slog.String("status", article.Status),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(article)
}

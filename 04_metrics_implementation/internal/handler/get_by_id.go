package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// GetArticle は記事取得のHTTPハンドラ
// GET /articles/{id}
func (h *ArticleHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	article, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get article",
			slog.String("id", id),
			slog.String("error", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	slog.InfoContext(ctx, "article retrieved",
		slog.String("id", article.ID),
		slog.String("title", article.Title),
		slog.String("status", article.Status),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(article)
}

package controller

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/tamaco489/otel_sample/04_metrics_implementation/internal/handler"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Server はHTTPサーバー
type Server struct {
	articleHandler *handler.ArticleHandler
	server         *http.Server
}

// NewServer は Server を生成
func NewServer(articleHandler *handler.ArticleHandler) *Server {
	return &Server{
		articleHandler: articleHandler,
	}
}

// Run はHTTPサーバーを起動
func (s *Server) Run(ctx context.Context, addr string) error {
	mux := http.NewServeMux()

	// ルーティング
	mux.HandleFunc("GET /articles/{id}", s.articleHandler.GetArticle)
	mux.HandleFunc("POST /articles", s.articleHandler.CreateArticle)

	// otelhttp でラップ (自動計装)
	otelHandler := otelhttp.NewHandler(mux, "http-server")

	s.server = &http.Server{
		Addr:    addr,
		Handler: otelHandler,
	}

	slog.InfoContext(ctx, "server starting", slog.String("addr", addr))
	return s.server.ListenAndServe()
}

// Shutdown はサーバーを停止
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

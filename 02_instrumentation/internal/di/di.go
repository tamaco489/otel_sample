package di

import (
	"github.com/tamaco489/otel_sample/02_instrumentation/internal/controller"
	"github.com/tamaco489/otel_sample/02_instrumentation/internal/handler"
	"github.com/tamaco489/otel_sample/02_instrumentation/internal/repository"
	"github.com/tamaco489/otel_sample/02_instrumentation/internal/usecase"
)

// Container は依存関係を保持するコンテナ
type Container struct {
	Server *controller.Server
}

// NewContainer は依存関係を初期化して Container を返す
func NewContainer() *Container {
	// Repository
	repo := repository.NewArticleRepository()

	// Usecase
	uc := usecase.NewArticleUsecase(repo)

	// Handler
	h := handler.NewArticleHandler(uc)

	// Controller
	srv := controller.NewServer(h)

	return &Container{
		Server: srv,
	}
}

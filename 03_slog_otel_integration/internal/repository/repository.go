package repository

import (
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("repository/article")

// Article は記事エンティティ
type Article struct {
	ID     string
	Title  string
	Status string
}

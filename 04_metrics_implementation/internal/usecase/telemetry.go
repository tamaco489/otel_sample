package usecase

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	tracer = otel.Tracer("usecase/article")
	meter  = otel.Meter("usecase/article")
)

var (
	articleViewsCounter   metric.Int64Counter
	articleCreateDuration metric.Float64Histogram
)

package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// OTELHandler は slog.Handler をラップし、trace_id / span_id を自動注入する
type OTELHandler struct {
	slog.Handler
}

// NewOTELHandler は OTELHandler を生成する
func NewOTELHandler(h slog.Handler) *OTELHandler {
	return &OTELHandler{Handler: h}
}

// Handle はログレコードに trace_id / span_id を追加してから内部ハンドラに委譲する
//
// NOTE: ロジック中の slog.InfoContext などが実行された場合、このメソッドが呼び出される
func (h *OTELHandler) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)
	// ctx から trace_id, span_id を抽出し、ログの構造体に追加
	if spanCtx.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs はラップされたハンドラに属性を追加した新しい OTELHandler を返す
func (h *OTELHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &OTELHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup はラップされたハンドラにグループを追加した新しい OTELHandler を返す
func (h *OTELHandler) WithGroup(name string) slog.Handler {
	return &OTELHandler{Handler: h.Handler.WithGroup(name)}
}

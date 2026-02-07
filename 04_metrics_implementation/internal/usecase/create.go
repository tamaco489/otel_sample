package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/tamaco489/otel_sample/04_metrics_implementation/internal/entity"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	apperrors "github.com/tamaco489/otel_sample/04_metrics_implementation/pkg/errors"
)

// CreateArticleInput は記事作成の入力
type CreateArticleInput struct {
	Title   string
	Content string
}

// Validate は入力を検証
func (i *CreateArticleInput) Validate() error {
	if i.Title == "" {
		return fmt.Errorf("%w: title is required", apperrors.ErrValidation)
	}
	return nil
}

// Create は記事を作成する (イベント記録の例)
func (u *articleUsecase) Create(ctx context.Context, input *CreateArticleInput) (*entity.Article, error) {
	// NOTE: Histogram (article.create.duration) の計測開始地点。
	// Histogram は値の分布を記録するメトリクスで、Record() に渡された値が事前に定義したバケット境界 (0.1, 0.5, 1, 2, 5秒) に振り分けられる。
	// 例: 0.3秒 → [0.1, 0.5) バケットに加算、2.1秒 → [2, 5) バケットに加算。
	// これにより「リクエストの何%が0.5秒以内に完了したか」等の分布分析が可能になる。
	// 成功・バリデーションエラー・DBエラーの全パスで status 属性付きで記録するため、関数冒頭で startTime を取得し、各 return 直前で Record() を呼ぶ。

	startTime := time.Now()

	ctx, span := tracer.Start(ctx, "ArticleUsecase.Create")
	defer span.End()

	// イベント: バリデーション開始
	span.AddEvent("validation_started")

	if err := input.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")

		// NOTE: Histogram 記録: バリデーションエラー時の処理時間
		articleCreateDuration.Record(ctx, time.Since(startTime).Seconds(),
			metric.WithAttributes(attribute.String("status", "validation_error")),
		)
		return nil, err
	}

	// イベント: バリデーション完了
	span.AddEvent("validation_completed")

	// イベント: 作成開始
	span.AddEvent("create_started")

	// リポジトリで作成
	article := &entity.Article{
		Title: input.Title,
	}
	created, err := u.repo.Create(ctx, article)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		// NOTE: Histogram 記録: DBエラー時の処理時間
		articleCreateDuration.Record(ctx, time.Since(startTime).Seconds(),
			metric.WithAttributes(attribute.String("status", "error")),
		)
		return nil, err
	}

	// イベント: 作成完了 (属性付き)
	span.AddEvent("create_completed", trace.WithAttributes(
		attribute.String("article.id", created.ID),
	))

	// NOTE: Histogram 記録: 成功時の処理時間
	articleCreateDuration.Record(ctx, time.Since(startTime).Seconds(),
		metric.WithAttributes(attribute.String("status", "success")),
	)

	return created, nil
}

package errors

import "errors"

var (
	// ErrNotFound はリソースが見つからないエラー
	ErrNotFound = errors.New("not found")

	// ErrValidation はバリデーションエラー
	ErrValidation = errors.New("validation error")
)

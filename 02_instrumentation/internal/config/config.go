package config

// Config はアプリケーション設定
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
}

// NewConfig はデフォルト設定を返す
func NewConfig() *Config {
	return &Config{
		ServiceName:    "article-api",
		ServiceVersion: "1.0.0",
		Environment:    "development",
	}
}

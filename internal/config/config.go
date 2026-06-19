package config

import "os"

type Config struct {
	ServiceName string
	Env         string
	HTTPAddr    string
	LogLevel    string
	LogFormat   string
	DBDSN       string
	SinkURL     string
}

func Load(service string) Config {
	return Config{
		ServiceName: Getenv("ALERTER_SERVICE", service),
		Env:         Getenv("ALERTER_ENV", "dev"),
		HTTPAddr:    Getenv("ALERTER_HTTP_ADDR", ":8080"),
		LogLevel:    Getenv("ALERTER_LOG_LEVEL", "info"),
		LogFormat:   Getenv("ALERTER_LOG_FORMAT", "json"),
		DBDSN:       Getenv("ALERTER_DB_DSN", ""),
		SinkURL:     Getenv("ALERTER_SINK_URL", ""),
	}
}

func Getenv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

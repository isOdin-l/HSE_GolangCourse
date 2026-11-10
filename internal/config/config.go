package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DbConfig
	ServerConfig
	TelemetryConfig
}

type TelemetryConfig struct {
	ServiceName      string        `env:"OTEL_SERVICE_NAME,notEmpty"`
	Endpoint         string        `env:"OTEL_EXPORTER_OTLP_ENDPOINT,notEmpty"`
	Protocol         string        `env:"OTEL_EXPORTER_OTLP_PROTOCOL,notEmpty"`
	ResourceAttrs    string        `env:"OTEL_RESOURCE_ATTRIBUTES"`
	TraceSampler     string        `env:"OTEL_TRACES_SAMPLER" envDefault:"parentbased_always_on"`
	TraceSamplerArgs string        `env:"OTEL_TRACES_SAMPLER_ARG"`
	FlushTimeout     time.Duration `env:"TELEMETRY_FLUSH_TIMEOUT" envDefault:"5s"`
}
type ServerConfig struct {
	HttpAddress       string        `env:"HTTP_ADDR,notEmpty"`
	LogLevel          string        `env:"LOG_LEVEL,notEmpty"`
	ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT,notEmpty"`
	ReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"5s"`
	ReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"2s"`
	WriteTimeout      time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
}

type DbConfig struct {
	Url             string        `env:"DATABASE_URL,notEmpty"`
	MaxConns        int32         `env:"DATABASE_MAX_CONNS,notEmpty"`
	MinConns        int32         `env:"DATABASE_MIN_CONNS,notEmpty"`
	ConnectTimeout  time.Duration `env:"DATABASE_CONNECT_TIMEOUT,notEmpty"`
	MaxConnLifeTime time.Duration `env:"DATABASE_MAX_CONN_LIFETIME,notEmpty"`
	QueryTimeout    time.Duration `env:"DATABASE_QUERY_TIMEOUT,notEmpty"` // для отмены по контексту
}

func Load() (Config, error) {
	return env.ParseAs[Config]()
}

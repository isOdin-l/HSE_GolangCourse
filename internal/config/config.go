package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DbConfig
	ServerConfig
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

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
	HttpAddress     string        `json:"HTTP_ADDR,notEmpty"`
	LogLevel        string        `json:"LOG_LEVEL,notEmpty"`
	ShutdownTimeout time.Duration `json:"SHUTDOWN_TIMEOUT,notEmpty"`
}

type DbConfig struct {
	Url             string        `json:"DATABASE_URL,notEmpty"`
	MaxConns        int32         `json:"DATABASE_MAX_CONNS,notEmpty"`
	MinConns        int32         `json:"DATABASE_MIN_CONNS,notEmpty"`
	ConnectTimeout  time.Duration `json:"DATABASE_CONNECT_TIMEOUT,notEmpty"`
	MaxConnLifeTime time.Duration `json:"DATABASE_MAX_CONN_LIFETIME,notEmpty"`
	QueryTimeout    time.Duration `json:"DATABASE_QUERY_TIMEOUT,notEmpty"` // для отмены по контексту
}

func Load() (Config, error) {
	return env.ParseAs[Config]()
}

package server

import (
	"net/http"

	"github.com/isOdin-l/HSE_GolangCourse.git/internal/config"
)

func New(cfg config.ServerConfig) *http.Server {
	return &http.Server{
		Addr: cfg.HttpAddress,
	}
}

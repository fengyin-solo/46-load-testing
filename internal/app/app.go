// Package app 负责依赖装配。
package app

import (
	"net/http"

	"loadtest/internal/config"
	"loadtest/internal/handler"
	"loadtest/internal/service"
	"loadtest/internal/store"
	"loadtest/pkg/logger"
)

// App 应用实例，聚合各层依赖。
type App struct {
	server *handler.Server
}

// New 装配应用依赖。
func New(cfg *config.Config, log *logger.Logger) (*App, error) {
	st := store.NewMemoryStore()
	svc := service.New(st, log, cfg)
	server := handler.NewServer(svc, log, cfg)
	log.Infof("应用装配完成，配置：%s", cfg.String())
	return &App{server: server}, nil
}

// Routes 返回根 HTTP 处理器。
func (a *App) Routes() http.Handler { return a.server.Routes() }

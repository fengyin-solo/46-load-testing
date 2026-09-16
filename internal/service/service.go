// Package service 实现压测平台的业务逻辑层。
package service

import (
	"loadtest/internal/config"
	"loadtest/internal/store"
	"loadtest/pkg/logger"
)

// Service 业务逻辑聚合根。
type Service struct {
	store store.Store
	log   *logger.Logger
	cfg   *config.Config
}

// New 构造 Service。
func New(st store.Store, log *logger.Logger, cfg *config.Config) *Service {
	return &Service{store: st, log: log, cfg: cfg}
}

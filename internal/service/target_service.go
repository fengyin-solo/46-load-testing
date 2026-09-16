package service

import (
	"sort"
	"time"

	"loadtest/internal/model"
	"loadtest/internal/store"
	"loadtest/pkg/idgen"
)

// CreateTarget 创建目标服务。
func (s *Service) CreateTarget(input model.Target) (*model.Target, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetTargetByURLMethod(input.URL, input.Method); err == nil {
		return nil, store.ErrConflict
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := s.store.CreateTarget(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建目标服务 %s(%s)", input.Name, input.ID)
	return &input, nil
}

// GetTarget 按 ID 查询目标服务。
func (s *Service) GetTarget(id string) (*model.Target, error) {
	return s.store.GetTarget(id)
}

// ListTargets 分页查询目标服务。
func (s *Service) ListTargets(filter model.TargetFilter, page, size int) ([]*model.Target, int, error) {
	all := s.store.ListTargets()
	matched := make([]*model.Target, 0, len(all))
	for _, t := range all {
		if filter.Match(t) {
			matched = append(matched, t)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Target{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateTarget 更新目标服务。
func (s *Service) UpdateTarget(id string, input model.Target) (*model.Target, error) {
	exist, err := s.store.GetTarget(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		exist.Name = input.Name
	}
	if input.URL != "" {
		exist.URL = input.URL
	}
	if input.Method != "" {
		exist.Method = input.Method
	}
	if input.Headers != nil {
		exist.Headers = input.Headers
	}
	if input.Status != "" {
		exist.Status = input.Status
	}
	if err := exist.Validate(); err != nil {
		return nil, err
	}
	exist.UpdatedAt = time.Now()
	if err := s.store.UpdateTarget(exist); err != nil {
		return nil, err
	}
	s.log.Infof("更新目标服务 %s(%s)", exist.Name, exist.ID)
	return exist, nil
}

// UpdateTargetStatus 更新目标服务健康状态。
func (s *Service) UpdateTargetStatus(id, status string) (*model.Target, error) {
	exist, err := s.store.GetTarget(id)
	if err != nil {
		return nil, err
	}
	if !model.IsTargetStatus(status) {
		return nil, model.NewValidationError("status", "目标服务状态不合法")
	}
	exist.Status = status
	exist.UpdatedAt = time.Now()
	if err := s.store.UpdateTarget(exist); err != nil {
		return nil, err
	}
	return exist, nil
}

// DeleteTarget 删除目标服务。
func (s *Service) DeleteTarget(id string) error {
	if _, err := s.store.GetTarget(id); err != nil {
		return err
	}
	return s.store.DeleteTarget(id)
}

package service

import (
	"sort"
	"time"

	"loadtest/internal/model"
	"loadtest/pkg/idgen"
)

// CreateExecutor 创建执行器。
func (s *Service) CreateExecutor(input model.Executor) (*model.Executor, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.LastHeartbeat = now
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := s.store.CreateExecutor(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建执行器 %s(%s)", input.Name, input.ID)
	return &input, nil
}

// GetExecutor 按 ID 查询执行器。
func (s *Service) GetExecutor(id string) (*model.Executor, error) {
	return s.store.GetExecutor(id)
}

// ListExecutors 分页查询执行器。
func (s *Service) ListExecutors(filter model.ExecutorFilter, page, size int) ([]*model.Executor, int, error) {
	all := s.store.ListExecutors()
	matched := make([]*model.Executor, 0, len(all))
	for _, e := range all {
		if filter.Match(e) {
			matched = append(matched, e)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Executor{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateExecutor 更新执行器。
func (s *Service) UpdateExecutor(id string, input model.Executor) (*model.Executor, error) {
	exist, err := s.store.GetExecutor(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		exist.Name = input.Name
	}
	if input.Addr != "" {
		exist.Addr = input.Addr
	}
	if input.Capacity > 0 {
		exist.Capacity = input.Capacity
	}
	if input.Status != "" {
		exist.Status = input.Status
	}
	if err := exist.Validate(); err != nil {
		return nil, err
	}
	exist.UpdatedAt = time.Now()
	if err := s.store.UpdateExecutor(exist); err != nil {
		return nil, err
	}
	s.log.Infof("更新执行器 %s(%s)", exist.Name, exist.ID)
	return exist, nil
}

// TransitionExecutor 执行执行器状态机流转。
func (s *Service) TransitionExecutor(id, to string) (*model.Executor, error) {
	exist, err := s.store.GetExecutor(id)
	if err != nil {
		return nil, err
	}
	if !model.IsExecutorStatus(to) {
		return nil, model.NewValidationError("status", "执行器状态不合法")
	}
	if !model.ExecutorCanTransition(exist.Status, to) {
		return nil, model.NewValidationError("status", "执行器状态不允许从 "+exist.Status+" 流转到 "+to)
	}
	exist.Status = to
	exist.UpdatedAt = time.Now()
	if err := s.store.UpdateExecutor(exist); err != nil {
		return nil, err
	}
	s.log.Infof("执行器 %s 状态流转 %s -> %s", exist.Name, exist.Status, to)
	return exist, nil
}

// Heartbeat 更新执行器心跳时间。
func (s *Service) Heartbeat(id string) (*model.Executor, error) {
	exist, err := s.store.GetExecutor(id)
	if err != nil {
		return nil, err
	}
	exist.LastHeartbeat = time.Now()
	exist.UpdatedAt = time.Now()
	if err := s.store.UpdateExecutor(exist); err != nil {
		return nil, err
	}
	return exist, nil
}

// DeleteExecutor 删除执行器。
func (s *Service) DeleteExecutor(id string) error {
	if _, err := s.store.GetExecutor(id); err != nil {
		return err
	}
	return s.store.DeleteExecutor(id)
}

package service

import (
	"sort"
	"time"

	"loadtest/internal/model"
	"loadtest/internal/store"
	"loadtest/pkg/idgen"
)

// CreateScenario 创建场景，校验目标服务存在与名称唯一。
func (s *Service) CreateScenario(input model.Scenario) (*model.Scenario, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetTarget(input.TargetID); err != nil {
		return nil, model.NewValidationError("target_id", "目标服务不存在")
	}
	if _, err := s.store.GetScenarioByName(input.Name); err == nil {
		return nil, store.ErrConflict
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := s.store.CreateScenario(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建场景 %s(%s)", input.Name, input.ID)
	return &input, nil
}

// GetScenario 按 ID 查询场景。
func (s *Service) GetScenario(id string) (*model.Scenario, error) {
	return s.store.GetScenario(id)
}

// ListScenarios 分页查询场景。
func (s *Service) ListScenarios(filter model.ScenarioFilter, page, size int) ([]*model.Scenario, int, error) {
	all := s.store.ListScenarios()
	matched := make([]*model.Scenario, 0, len(all))
	for _, sc := range all {
		if filter.Match(sc) {
			matched = append(matched, sc)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Scenario{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateScenario 更新场景。
func (s *Service) UpdateScenario(id string, input model.Scenario) (*model.Scenario, error) {
	exist, err := s.store.GetScenario(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		exist.Name = input.Name
	}
	if input.TargetID != "" {
		exist.TargetID = input.TargetID
	}
	if input.Method != "" {
		exist.Method = input.Method
	}
	if input.Path != "" {
		exist.Path = input.Path
	}
	if input.Concurrent > 0 {
		exist.Concurrent = input.Concurrent
	}
	if input.DurationSec > 0 {
		exist.DurationSec = input.DurationSec
	}
	if input.RampUpSec >= 0 {
		exist.RampUpSec = input.RampUpSec
	}
	if input.Body != "" {
		exist.Body = input.Body
	}
	if input.Status != "" {
		exist.Status = input.Status
	}
	if err := exist.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetTarget(exist.TargetID); err != nil {
		return nil, model.NewValidationError("target_id", "目标服务不存在")
	}
	exist.UpdatedAt = time.Now()
	if err := s.store.UpdateScenario(exist); err != nil {
		return nil, err
	}
	s.log.Infof("更新场景 %s(%s)", exist.Name, exist.ID)
	return exist, nil
}

// TransitionScenario 执行场景状态机流转。
func (s *Service) TransitionScenario(id, to string) (*model.Scenario, error) {
	exist, err := s.store.GetScenario(id)
	if err != nil {
		return nil, err
	}
	if !model.IsScenarioStatus(to) {
		return nil, model.NewValidationError("status", "场景状态不合法")
	}
	if !model.ScenarioCanTransition(exist.Status, to) {
		return nil, model.NewValidationError("status", "场景状态不允许从 "+exist.Status+" 流转到 "+to)
	}
	exist.Status = to
	exist.UpdatedAt = time.Now()
	if err := s.store.UpdateScenario(exist); err != nil {
		return nil, err
	}
	s.log.Infof("场景 %s 状态流转 %s -> %s", exist.Name, exist.Status, to)
	return exist, nil
}

// DeleteScenario 删除场景。
func (s *Service) DeleteScenario(id string) error {
	if _, err := s.store.GetScenario(id); err != nil {
		return err
	}
	return s.store.DeleteScenario(id)
}

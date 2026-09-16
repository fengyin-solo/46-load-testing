package service

import (
	"sort"
	"time"

	"loadtest/internal/model"
	"loadtest/pkg/idgen"
)

// CreateTestPlan 创建测试计划，校验场景存在。
func (s *Service) CreateTestPlan(input model.TestPlan) (*model.TestPlan, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetScenario(input.ScenarioID); err != nil {
		return nil, model.NewValidationError("scenario_id", "场景不存在")
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	input.UpdatedAt = now
	if err := s.store.CreateTestPlan(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建测试计划 %s(%s)", input.Name, input.ID)
	return &input, nil
}

// GetTestPlan 按 ID 查询测试计划。
func (s *Service) GetTestPlan(id string) (*model.TestPlan, error) {
	return s.store.GetTestPlan(id)
}

// ListTestPlans 分页查询测试计划。
func (s *Service) ListTestPlans(filter model.PlanFilter, page, size int) ([]*model.TestPlan, int, error) {
	all := s.store.ListTestPlans()
	matched := make([]*model.TestPlan, 0, len(all))
	for _, p := range all {
		if filter.Match(p) {
			matched = append(matched, p)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.TestPlan{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateTestPlan 更新测试计划。
func (s *Service) UpdateTestPlan(id string, input model.TestPlan) (*model.TestPlan, error) {
	exist, err := s.store.GetTestPlan(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		exist.Name = input.Name
	}
	if input.ScenarioID != "" {
		exist.ScenarioID = input.ScenarioID
	}
	if input.Description != "" {
		exist.Description = input.Description
	}
	if input.Status != "" {
		exist.Status = input.Status
	}
	if err := exist.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetScenario(exist.ScenarioID); err != nil {
		return nil, model.NewValidationError("scenario_id", "场景不存在")
	}
	exist.UpdatedAt = time.Now()
	if err := s.store.UpdateTestPlan(exist); err != nil {
		return nil, err
	}
	s.log.Infof("更新测试计划 %s(%s)", exist.Name, exist.ID)
	return exist, nil
}

// TransitionPlan 执行测试计划状态机流转。
func (s *Service) TransitionPlan(id, to string) (*model.TestPlan, error) {
	exist, err := s.store.GetTestPlan(id)
	if err != nil {
		return nil, err
	}
	if !model.IsPlanStatus(to) {
		return nil, model.NewValidationError("status", "测试计划状态不合法")
	}
	if !model.PlanCanTransition(exist.Status, to) {
		return nil, model.NewValidationError("status", "测试计划状态不允许从 "+exist.Status+" 流转到 "+to)
	}
	exist.Status = to
	exist.UpdatedAt = time.Now()
	if err := s.store.UpdateTestPlan(exist); err != nil {
		return nil, err
	}
	s.log.Infof("测试计划 %s 状态流转 %s -> %s", exist.Name, exist.Status, to)
	return exist, nil
}

// BatchTransitionPlans 批量更新测试计划状态。
func (s *Service) BatchTransitionPlans(ids []string, to string) ([]*model.TestPlan, error) {
	if !model.IsPlanStatus(to) {
		return nil, model.NewValidationError("status", "测试计划状态不合法")
	}
	updated := make([]*model.TestPlan, 0, len(ids))
	for _, id := range ids {
		p, err := s.store.GetTestPlan(id)
		if err != nil {
			return nil, err
		}
		if !model.PlanCanTransition(p.Status, to) {
			return nil, model.NewValidationError("status", "测试计划 "+p.Name+" 状态不允许从 "+p.Status+" 流转到 "+to)
		}
		p.Status = to
		p.UpdatedAt = time.Now()
		if err := s.store.UpdateTestPlan(p); err != nil {
			return nil, err
		}
		updated = append(updated, p)
	}
	s.log.Infof("批量更新 %d 个测试计划状态为 %s", len(updated), to)
	return updated, nil
}

// DeleteTestPlan 删除测试计划。
func (s *Service) DeleteTestPlan(id string) error {
	if _, err := s.store.GetTestPlan(id); err != nil {
		return err
	}
	return s.store.DeleteTestPlan(id)
}

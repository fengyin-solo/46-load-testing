package store

import (
	"loadtest/internal/model"
)

// CreateTestPlan 创建测试计划。
func (s *MemoryStore) CreateTestPlan(p *model.TestPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.testplans[p.ID] = p
	return nil
}

// GetTestPlan 按 ID 查询测试计划。
func (s *MemoryStore) GetTestPlan(id string) (*model.TestPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.testplans[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

// ListTestPlans 返回全部测试计划。
func (s *MemoryStore) ListTestPlans() []*model.TestPlan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.TestPlan, 0, len(s.testplans))
	for _, p := range s.testplans {
		list = append(list, p)
	}
	return list
}

// UpdateTestPlan 更新测试计划。
func (s *MemoryStore) UpdateTestPlan(p *model.TestPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.testplans[p.ID]; !ok {
		return ErrNotFound
	}
	s.testplans[p.ID] = p
	return nil
}

// DeleteTestPlan 删除测试计划。
func (s *MemoryStore) DeleteTestPlan(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.testplans[id]; !ok {
		return ErrNotFound
	}
	delete(s.testplans, id)
	return nil
}

package store

import (
	"loadtest/internal/model"
)

// CreateScenario 创建场景，校验名称唯一性。
func (s *MemoryStore) CreateScenario(sc *model.Scenario) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.scenarios {
		if exist.Name == sc.Name {
			return ErrConflict
		}
	}
	s.scenarios[sc.ID] = sc
	return nil
}

// GetScenario 按 ID 查询场景。
func (s *MemoryStore) GetScenario(id string) (*model.Scenario, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok := s.scenarios[id]
	if !ok {
		return nil, ErrNotFound
	}
	return sc, nil
}

// GetScenarioByName 按名称查询场景。
func (s *MemoryStore) GetScenarioByName(name string) (*model.Scenario, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sc := range s.scenarios {
		if sc.Name == name {
			return sc, nil
		}
	}
	return nil, ErrNotFound
}

// ListScenarios 返回全部场景。
func (s *MemoryStore) ListScenarios() []*model.Scenario {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Scenario, 0, len(s.scenarios))
	for _, sc := range s.scenarios {
		list = append(list, sc)
	}
	return list
}

// UpdateScenario 更新场景。
func (s *MemoryStore) UpdateScenario(sc *model.Scenario) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.scenarios[sc.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.scenarios {
		if exist.ID != sc.ID && exist.Name == sc.Name {
			return ErrConflict
		}
	}
	s.scenarios[sc.ID] = sc
	return nil
}

// DeleteScenario 删除场景。
func (s *MemoryStore) DeleteScenario(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.scenarios[id]; !ok {
		return ErrNotFound
	}
	delete(s.scenarios, id)
	return nil
}

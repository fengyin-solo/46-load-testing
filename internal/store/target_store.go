package store

import (
	"loadtest/internal/model"
)

// CreateTarget 创建目标服务，校验 URL+Method 唯一性。
func (s *MemoryStore) CreateTarget(t *model.Target) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.targets {
		if exist.URL == t.URL && exist.Method == t.Method {
			return ErrConflict
		}
	}
	s.targets[t.ID] = t
	return nil
}

// GetTarget 按 ID 查询目标服务。
func (s *MemoryStore) GetTarget(id string) (*model.Target, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.targets[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

// GetTargetByURLMethod 按 URL+Method 查询目标服务。
func (s *MemoryStore) GetTargetByURLMethod(url, method string) (*model.Target, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.targets {
		if t.URL == url && t.Method == method {
			return t, nil
		}
	}
	return nil, ErrNotFound
}

// ListTargets 返回全部目标服务。
func (s *MemoryStore) ListTargets() []*model.Target {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Target, 0, len(s.targets))
	for _, t := range s.targets {
		list = append(list, t)
	}
	return list
}

// UpdateTarget 更新目标服务。
func (s *MemoryStore) UpdateTarget(t *model.Target) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.targets[t.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.targets {
		if exist.ID != t.ID && exist.URL == t.URL && exist.Method == t.Method {
			return ErrConflict
		}
	}
	s.targets[t.ID] = t
	return nil
}

// DeleteTarget 删除目标服务。
func (s *MemoryStore) DeleteTarget(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.targets[id]; !ok {
		return ErrNotFound
	}
	delete(s.targets, id)
	return nil
}

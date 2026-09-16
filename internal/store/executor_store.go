package store

import (
	"loadtest/internal/model"
)

// CreateExecutor 创建执行器，校验名称唯一性。
func (s *MemoryStore) CreateExecutor(e *model.Executor) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.executors {
		if exist.Name == e.Name {
			return ErrConflict
		}
	}
	s.executors[e.ID] = e
	return nil
}

// GetExecutor 按 ID 查询执行器。
func (s *MemoryStore) GetExecutor(id string) (*model.Executor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.executors[id]
	if !ok {
		return nil, ErrNotFound
	}
	return e, nil
}

// ListExecutors 返回全部执行器。
func (s *MemoryStore) ListExecutors() []*model.Executor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Executor, 0, len(s.executors))
	for _, e := range s.executors {
		list = append(list, e)
	}
	return list
}

// UpdateExecutor 更新执行器。
func (s *MemoryStore) UpdateExecutor(e *model.Executor) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.executors[e.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.executors {
		if exist.ID != e.ID && exist.Name == e.Name {
			return ErrConflict
		}
	}
	s.executors[e.ID] = e
	return nil
}

// DeleteExecutor 删除执行器。
func (s *MemoryStore) DeleteExecutor(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.executors[id]; !ok {
		return ErrNotFound
	}
	delete(s.executors, id)
	return nil
}

package store

import (
	"loadtest/internal/model"
)

// CreateSchedule 创建调度。
func (s *MemoryStore) CreateSchedule(sc *model.Schedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules[sc.ID] = sc
	return nil
}

// GetSchedule 按 ID 查询调度。
func (s *MemoryStore) GetSchedule(id string) (*model.Schedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok := s.schedules[id]
	if !ok {
		return nil, ErrNotFound
	}
	return sc, nil
}

// ListSchedules 返回全部调度。
func (s *MemoryStore) ListSchedules() []*model.Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Schedule, 0, len(s.schedules))
	for _, sc := range s.schedules {
		list = append(list, sc)
	}
	return list
}

// UpdateSchedule 更新调度。
func (s *MemoryStore) UpdateSchedule(sc *model.Schedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.schedules[sc.ID]; !ok {
		return ErrNotFound
	}
	s.schedules[sc.ID] = sc
	return nil
}

// DeleteSchedule 删除调度。
func (s *MemoryStore) DeleteSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.schedules[id]; !ok {
		return ErrNotFound
	}
	delete(s.schedules, id)
	return nil
}

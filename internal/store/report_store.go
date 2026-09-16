package store

import (
	"loadtest/internal/model"
)

// CreateReport 创建报告，校验执行记录唯一性。
func (s *MemoryStore) CreateReport(r *model.Report) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.reports {
		if exist.ExecutionID == r.ExecutionID {
			return ErrConflict
		}
	}
	s.reports[r.ID] = r
	return nil
}

// GetReport 按 ID 查询报告。
func (s *MemoryStore) GetReport(id string) (*model.Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.reports[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}

// GetReportByExecution 按执行记录 ID 查询报告。
func (s *MemoryStore) GetReportByExecution(executionID string) (*model.Report, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.reports {
		if r.ExecutionID == executionID {
			return r, nil
		}
	}
	return nil, ErrNotFound
}

// ListReports 返回全部报告。
func (s *MemoryStore) ListReports() []*model.Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Report, 0, len(s.reports))
	for _, r := range s.reports {
		list = append(list, r)
	}
	return list
}

// DeleteReport 删除报告。
func (s *MemoryStore) DeleteReport(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.reports[id]; !ok {
		return ErrNotFound
	}
	delete(s.reports, id)
	return nil
}

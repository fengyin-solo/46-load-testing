package store

import (
	"loadtest/internal/model"
)

// CreateMetricSample 创建单条指标样本。
func (s *MemoryStore) CreateMetricSample(m *model.MetricSample) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics[m.ID] = m
	return nil
}

// CreateMetricSamples 批量创建指标样本。
func (s *MemoryStore) CreateMetricSamples(ms []*model.MetricSample) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range ms {
		s.metrics[m.ID] = m
	}
	return nil
}

// GetMetricSample 按 ID 查询指标样本。
func (s *MemoryStore) GetMetricSample(id string) (*model.MetricSample, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.metrics[id]
	if !ok {
		return nil, ErrNotFound
	}
	return m, nil
}

// ListMetricSamples 返回全部指标样本。
func (s *MemoryStore) ListMetricSamples() []*model.MetricSample {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.MetricSample, 0, len(s.metrics))
	for _, m := range s.metrics {
		list = append(list, m)
	}
	return list
}

// DeleteMetricSample 删除单条指标样本。
func (s *MemoryStore) DeleteMetricSample(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.metrics[id]; !ok {
		return ErrNotFound
	}
	delete(s.metrics, id)
	return nil
}

// DeleteMetricSamplesByExecution 批量删除某执行记录的全部指标样本。
func (s *MemoryStore) DeleteMetricSamplesByExecution(executionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, m := range s.metrics {
		if m.ExecutionID == executionID {
			delete(s.metrics, id)
		}
	}
	return nil
}

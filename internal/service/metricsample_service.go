package service

import (
	"sort"
	"time"

	"loadtest/internal/model"
	"loadtest/pkg/idgen"
)

// CreateMetricSample 创建单条指标样本。
func (s *Service) CreateMetricSample(input model.MetricSample) (*model.MetricSample, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetExecution(input.ExecutionID); err != nil {
		return nil, model.NewValidationError("execution_id", "执行记录不存在")
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	if err := s.store.CreateMetricSample(&input); err != nil {
		return nil, err
	}
	return &input, nil
}

// CreateMetricSamples 批量创建指标样本。
func (s *Service) CreateMetricSamples(executionID string, inputs []model.MetricSample) ([]*model.MetricSample, error) {
	if _, err := s.store.GetExecution(executionID); err != nil {
		return nil, model.NewValidationError("execution_id", "执行记录不存在")
	}
	now := time.Now()
	created := make([]*model.MetricSample, 0, len(inputs))
	for i := range inputs {
		inputs[i].ExecutionID = executionID
		if inputs[i].Timestamp.IsZero() {
			inputs[i].Timestamp = now
		}
		if err := inputs[i].Validate(); err != nil {
			return nil, err
		}
		inputs[i].ID = idgen.Hex()
		inputs[i].CreatedAt = now
		created = append(created, &inputs[i])
	}
	if err := s.store.CreateMetricSamples(created); err != nil {
		return nil, err
	}
	s.log.Infof("批量写入 %d 条指标样本(执行=%s)", len(created), executionID)
	return created, nil
}

// GetMetricSample 按 ID 查询指标样本。
func (s *Service) GetMetricSample(id string) (*model.MetricSample, error) {
	return s.store.GetMetricSample(id)
}

// ListMetricSamples 分页查询指标样本。
func (s *Service) ListMetricSamples(filter model.MetricSampleFilter, page, size int) ([]*model.MetricSample, int, error) {
	all := s.store.ListMetricSamples()
	matched := make([]*model.MetricSample, 0, len(all))
	for _, m := range all {
		if filter.Match(m) {
			matched = append(matched, m)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Timestamp.After(matched[j].Timestamp)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.MetricSample{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// AggregateByExecution 聚合某执行记录的指标样本。
func (s *Service) AggregateByExecution(executionID string) (*model.MetricAggregation, error) {
	if _, err := s.store.GetExecution(executionID); err != nil {
		return nil, err
	}
	all := s.store.ListMetricSamples()
	samples := make(model.MetricSamples, 0, len(all))
	for _, m := range all {
		if m.ExecutionID == executionID {
			samples = append(samples, m)
		}
	}
	agg := &model.MetricAggregation{
		ExecutionID: executionID,
		SampleCount: len(samples),
		AvgTPS:      samples.AvgTPS(),
		AvgLatency:  samples.AvgLatency(),
		MaxP99:      samples.MaxP99(),
		AvgErrorRate: samples.AvgErrorRate(),
	}
	return agg, nil
}

// DeleteMetricSample 删除指标样本。
func (s *Service) DeleteMetricSample(id string) error {
	if _, err := s.store.GetMetricSample(id); err != nil {
		return err
	}
	return s.store.DeleteMetricSample(id)
}

package service

import (
	"sort"
	"time"

	"loadtest/internal/model"
	"loadtest/pkg/idgen"
)

// CreateAssertion 创建断言，校验执行记录存在并立即判定结果。
func (s *Service) CreateAssertion(input model.Assertion) (*model.Assertion, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetExecution(input.ExecutionID); err != nil {
		return nil, model.NewValidationError("execution_id", "执行记录不存在")
	}
	actual, err := s.aggregateMetricForAssertion(input.ExecutionID, input.Metric)
	if err != nil {
		return nil, err
	}
	if input.Evaluate(actual) {
		input.Result = model.AssertionPass
	} else {
		input.Result = model.AssertionFail
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	if err := s.store.CreateAssertion(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建断言 %s(执行=%s, 指标=%s, 结果=%s)", input.ID, input.ExecutionID, input.Metric, input.Result)
	return &input, nil
}

// aggregateMetricForAssertion 根据断言指标聚合执行记录的实际值。
func (s *Service) aggregateMetricForAssertion(executionID, metric string) (float64, error) {
	agg, err := s.AggregateByExecution(executionID)
	if err != nil {
		return 0, err
	}
	switch metric {
	case model.MetricTPS:
		return agg.AvgTPS, nil
	case model.MetricAvgLatency:
		return agg.AvgLatency, nil
	case model.MetricP99:
		return agg.MaxP99, nil
	case model.MetricErrorRate:
		return agg.AvgErrorRate, nil
	default:
		return 0, model.NewValidationError("metric", "断言指标不合法")
	}
}

// GetAssertion 按 ID 查询断言。
func (s *Service) GetAssertion(id string) (*model.Assertion, error) {
	return s.store.GetAssertion(id)
}

// ListAssertions 分页查询断言。
func (s *Service) ListAssertions(filter model.AssertionFilter, page, size int) ([]*model.Assertion, int, error) {
	all := s.store.ListAssertions()
	matched := make([]*model.Assertion, 0, len(all))
	for _, a := range all {
		if filter.Match(a) {
			matched = append(matched, a)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Assertion{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// DeleteAssertion 删除断言。
func (s *Service) DeleteAssertion(id string) error {
	if _, err := s.store.GetAssertion(id); err != nil {
		return err
	}
	return s.store.DeleteAssertion(id)
}

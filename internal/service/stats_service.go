package service

import (
	"sort"
	"time"

	"loadtest/internal/model"
)

// Overview 生成全局统计快照。
func (s *Service) Overview() (*model.OverviewStats, error) {
	targets := s.store.ListTargets()
	scenarios := s.store.ListScenarios()
	plans := s.store.ListTestPlans()
	executors := s.store.ListExecutors()
	executions := s.store.ListExecutions()
	metrics := s.store.ListMetricSamples()
	reports := s.store.ListReports()
	schedules := s.store.ListSchedules()

	stats := &model.OverviewStats{
		TargetCount:    len(targets),
		ScenarioCount:  len(scenarios),
		PlanCount:      len(plans),
		ExecutorCount:  len(executors),
		ExecutionCount: len(executions),
		MetricCount:    len(metrics),
		ReportCount:    len(reports),
		ScheduleCount:  len(schedules),
		ByStatus:       make(map[string]int),
		GeneratedAt:    time.Now(),
	}
	var totalRequests, totalErrors int64
	for _, e := range executions {
		stats.ByStatus[e.Status]++
		if e.Status == model.ExecutionRunning {
			stats.RunningExecutions++
		}
		totalRequests += e.TotalRequests
		totalErrors += e.ErrorCount
	}
	stats.TotalRequests = totalRequests
	stats.TotalErrors = totalErrors
	if totalRequests > 0 {
		stats.GlobalErrorRate = float64(totalErrors) / float64(totalRequests)
	}
	samples := model.MetricSamples(metrics)
	stats.GlobalAvgTPS = samples.AvgTPS()
	stats.GlobalAvgLatency = samples.AvgLatency()
	stats.GlobalMaxP99 = samples.MaxP99()
	return stats, nil
}

// AggregateExecution 聚合单个执行记录的指标与统计。
func (s *Service) AggregateExecution(executionID string) (*model.ExecutionAggregation, error) {
	execution, err := s.store.GetExecution(executionID)
	if err != nil {
		return nil, err
	}
	agg, err := s.AggregateByExecution(executionID)
	if err != nil {
		return nil, err
	}
	return &model.ExecutionAggregation{
		ExecutionID:   execution.ID,
		PlanID:        execution.PlanID,
		Status:        execution.Status,
		TotalRequests: execution.TotalRequests,
		ErrorCount:    execution.ErrorCount,
		ErrorRate:     execution.ErrorRate(),
		AvgTPS:        agg.AvgTPS,
		AvgLatency:    agg.AvgLatency,
		MaxP99:        agg.MaxP99,
		SampleCount:   agg.SampleCount,
	}, nil
}

// AggregationsByPlan 按测试计划聚合压测结果。
func (s *Service) AggregationsByPlan() ([]*model.PlanAggregation, error) {
	plans := s.store.ListTestPlans()
	planNames := make(map[string]string, len(plans))
	for _, p := range plans {
		planNames[p.ID] = p.Name
	}
	executions := s.store.ListExecutions()
	metrics := s.store.ListMetricSamples()
	metricsByExec := make(map[string]model.MetricSamples)
	for _, m := range metrics {
		metricsByExec[m.ExecutionID] = append(metricsByExec[m.ExecutionID], m)
	}

	byPlan := make(map[string]*model.PlanAggregation)
	for _, e := range executions {
		pa, ok := byPlan[e.PlanID]
		if !ok {
			pa = &model.PlanAggregation{PlanID: e.PlanID, PlanName: planNames[e.PlanID]}
			byPlan[e.PlanID] = pa
		}
		pa.ExecutionCount++
		pa.TotalRequests += e.TotalRequests
		pa.ErrorCount += e.ErrorCount
		ms := metricsByExec[e.ID]
		pa.AvgTPS += ms.AvgTPS()
		pa.AvgLatency += ms.AvgLatency()
		if max := ms.MaxP99(); max > pa.MaxP99 {
			pa.MaxP99 = max
		}
	}
	result := make([]*model.PlanAggregation, 0, len(byPlan))
	for _, pa := range byPlan {
		if pa.TotalRequests > 0 {
			pa.ErrorRate = float64(pa.ErrorCount) / float64(pa.TotalRequests)
		}
		if pa.ExecutionCount > 0 {
			pa.AvgTPS /= float64(pa.ExecutionCount)
			pa.AvgLatency /= float64(pa.ExecutionCount)
		}
		result = append(result, pa)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalRequests > result[j].TotalRequests
	})
	return result, nil
}

// AggregationsByExecutor 按执行器聚合压测结果。
func (s *Service) AggregationsByExecutor() ([]*model.ExecutorAggregation, error) {
	executors := s.store.ListExecutors()
	executorNames := make(map[string]string, len(executors))
	for _, e := range executors {
		executorNames[e.ID] = e.Name
	}
	executions := s.store.ListExecutions()
	byExecutor := make(map[string]*model.ExecutorAggregation)
	for _, e := range executions {
		ea, ok := byExecutor[e.ExecutorID]
		if !ok {
			ea = &model.ExecutorAggregation{ExecutorID: e.ExecutorID, ExecutorName: executorNames[e.ExecutorID]}
			byExecutor[e.ExecutorID] = ea
		}
		ea.ExecutionCount++
		ea.TotalRequests += e.TotalRequests
		ea.ErrorCount += e.ErrorCount
	}
	result := make([]*model.ExecutorAggregation, 0, len(byExecutor))
	for _, ea := range byExecutor {
		if ea.TotalRequests > 0 {
			ea.ErrorRate = float64(ea.ErrorCount) / float64(ea.TotalRequests)
		}
		result = append(result, ea)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ExecutionCount > result[j].ExecutionCount
	})
	return result, nil
}

// TopScenarios 按场景聚合执行记录并返回 TOP N。
func (s *Service) TopScenarios(n int) ([]*model.TopScenario, error) {
	if n <= 0 {
		n = 10
	}
	scenarios := s.store.ListScenarios()
	scenarioNames := make(map[string]string, len(scenarios))
	for _, sc := range scenarios {
		scenarioNames[sc.ID] = sc.Name
	}
	plans := s.store.ListTestPlans()
	planScenario := make(map[string]string, len(plans))
	for _, p := range plans {
		planScenario[p.ID] = p.ScenarioID
	}
	executions := s.store.ListExecutions()
	metrics := s.store.ListMetricSamples()
	metricsByExec := make(map[string]model.MetricSamples)
	for _, m := range metrics {
		metricsByExec[m.ExecutionID] = append(metricsByExec[m.ExecutionID], m)
	}

	byScenario := make(map[string]*model.TopScenario)
	for _, e := range executions {
		scenarioID := planScenario[e.PlanID]
		ts, ok := byScenario[scenarioID]
		if !ok {
			ts = &model.TopScenario{ScenarioID: scenarioID, ScenarioName: scenarioNames[scenarioID]}
			byScenario[scenarioID] = ts
		}
		ts.ExecutionCount++
		ts.TotalRequests += e.TotalRequests
		ts.AvgTPS += metricsByExec[e.ID].AvgTPS()
	}
	result := make([]*model.TopScenario, 0, len(byScenario))
	for _, ts := range byScenario {
		if ts.ExecutionCount > 0 {
			ts.AvgTPS /= float64(ts.ExecutionCount)
		}
		result = append(result, ts)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalRequests > result[j].TotalRequests
	})
	if len(result) > n {
		result = result[:n]
	}
	return result, nil
}

// ExportSnapshot 导出汇总快照。
func (s *Service) ExportSnapshot(topN int) (*model.ExportSnapshot, error) {
	overview, err := s.Overview()
	if err != nil {
		return nil, err
	}
	plans, err := s.AggregationsByPlan()
	if err != nil {
		return nil, err
	}
	executors, err := s.AggregationsByExecutor()
	if err != nil {
		return nil, err
	}
	top, err := s.TopScenarios(topN)
	if err != nil {
		return nil, err
	}
	return &model.ExportSnapshot{
		GeneratedAt:  time.Now(),
		Overview:     *overview,
		Plans:        plans,
		Executors:    executors,
		TopScenarios: top,
	}, nil
}

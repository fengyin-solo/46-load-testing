package service

import (
	"fmt"
	"sort"
	"time"

	"loadtest/internal/model"
	"loadtest/pkg/idgen"
)

// GenerateReport 聚合执行记录指标与断言，生成压测报告。
func (s *Service) GenerateReport(executionID string) (*model.Report, error) {
	execution, err := s.store.GetExecution(executionID)
	if err != nil {
		return nil, err
	}
	if !model.ExecutionIsFinal(execution.Status) {
		return nil, model.NewValidationError("status", "执行记录未完成，无法生成报告")
	}
	agg, err := s.AggregateByExecution(executionID)
	if err != nil {
		return nil, err
	}
	all := s.store.ListAssertions()
	assertions := make([]*model.Assertion, 0, len(all))
	passCount, failCount := 0, 0
	for _, a := range all {
		if a.ExecutionID == executionID {
			assertions = append(assertions, a)
			if a.Result == model.AssertionPass {
				passCount++
			} else if a.Result == model.AssertionFail {
				failCount++
			}
		}
	}
	summary := buildReportSummary(execution, agg, assertions, passCount, failCount)
	conclusion := buildReportConclusion(execution, failCount)
	report := &model.Report{
		ID:          idgen.Hex(),
		ExecutionID: executionID,
		Summary:     summary,
		Conclusion:  conclusion,
		GeneratedAt: time.Now(),
		CreatedAt:   time.Now(),
	}
	if err := s.store.CreateReport(report); err != nil {
		return nil, err
	}
	s.log.Infof("生成报告 %s(执行=%s)", report.ID, executionID)
	return report, nil
}

// buildReportSummary 构造报告摘要文本。
func buildReportSummary(e *model.Execution, agg *model.MetricAggregation, assertions []*model.Assertion, pass, fail int) string {
	status := map[string]string{
		model.ExecutionCompleted: "完成",
		model.ExecutionFailed:    "失败",
		model.ExecutionCancelled: "取消",
	}[e.Status]
	started := "-"
	if e.StartedAt != nil {
		started = e.StartedAt.Format(time.RFC3339)
	}
	finished := "-"
	if e.FinishedAt != nil {
		finished = e.FinishedAt.Format(time.RFC3339)
	}
	return fmt.Sprintf(
		"执行记录 %s 状态为 %s。开始时间 %s，结束时间 %s。共发起 %d 个请求，错误 %d 个，错误率 %.2f%%。指标样本 %d 条：平均 TPS %.2f，平均延迟 %.2f ms，P99 峰值 %.2f ms，平均错误率 %.2f%%。断言共 %d 条，通过 %d 条，失败 %d 条。",
		e.ID, status, started, finished, e.TotalRequests, e.ErrorCount,
		e.ErrorRate()*100, agg.SampleCount, agg.AvgTPS, agg.AvgLatency, agg.MaxP99,
		agg.AvgErrorRate*100, len(assertions), pass, fail,
	)
}

// buildReportConclusion 构造报告结论文本。
func buildReportConclusion(e *model.Execution, failCount int) string {
	if e.Status == model.ExecutionFailed {
		return "本次压测执行失败，建议检查压测目标稳定性与执行器资源后重试。"
	}
	if e.Status == model.ExecutionCancelled {
		return "本次压测已取消，未产生完整的结论。"
	}
	if failCount > 0 {
		return fmt.Sprintf("本次压测存在 %d 条断言未通过，系统性能未达到预期阈值，建议优化后复测。", failCount)
	}
	return "本次压测全部断言通过，系统性能满足预期阈值，压测结论为通过。"
}

// GetReport 按 ID 查询报告。
func (s *Service) GetReport(id string) (*model.Report, error) {
	return s.store.GetReport(id)
}

// GetReportByExecution 按执行记录查询报告。
func (s *Service) GetReportByExecution(executionID string) (*model.Report, error) {
	return s.store.GetReportByExecution(executionID)
}

// ListReports 分页查询报告。
func (s *Service) ListReports(filter model.ReportFilter, page, size int) ([]*model.Report, int, error) {
	all := s.store.ListReports()
	matched := make([]*model.Report, 0, len(all))
	for _, r := range all {
		if filter.Match(r) {
			matched = append(matched, r)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].GeneratedAt.After(matched[j].GeneratedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Report{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// DeleteReport 删除报告。
func (s *Service) DeleteReport(id string) error {
	if _, err := s.store.GetReport(id); err != nil {
		return err
	}
	return s.store.DeleteReport(id)
}

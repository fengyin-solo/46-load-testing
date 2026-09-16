package service

import (
	"sort"
	"time"

	"loadtest/internal/model"
	"loadtest/pkg/idgen"
)

// CreateExecution 创建执行记录，校验计划与执行器存在且执行器在线。
func (s *Service) CreateExecution(input model.Execution) (*model.Execution, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetTestPlan(input.PlanID); err != nil {
		return nil, model.NewValidationError("plan_id", "测试计划不存在")
	}
	executor, err := s.store.GetExecutor(input.ExecutorID)
	if err != nil {
		return nil, model.NewValidationError("executor_id", "执行器不存在")
	}
	if executor.Status != model.ExecutorOnline {
		return nil, model.NewValidationError("executor_id", "执行器不在线，无法分配执行")
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.Status = model.ExecutionPending
	input.CreatedAt = now
	if err := s.store.CreateExecution(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建执行记录 %s(计划=%s)", input.ID, input.PlanID)
	return &input, nil
}

// GetExecution 按 ID 查询执行记录。
func (s *Service) GetExecution(id string) (*model.Execution, error) {
	return s.store.GetExecution(id)
}

// ListExecutions 分页查询执行记录。
func (s *Service) ListExecutions(filter model.ExecutionFilter, page, size int) ([]*model.Execution, int, error) {
	all := s.store.ListExecutions()
	matched := make([]*model.Execution, 0, len(all))
	for _, e := range all {
		if filter.Match(e) {
			matched = append(matched, e)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Execution{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// StartExecution 启动执行：pending -> running，执行器置为 busy。
func (s *Service) StartExecution(id string) (*model.Execution, error) {
	exist, err := s.store.GetExecution(id)
	if err != nil {
		return nil, err
	}
	if !model.ExecutionCanTransition(exist.Status, model.ExecutionRunning) {
		return nil, model.NewValidationError("status", "执行记录不允许从 "+exist.Status+" 流转到 running")
	}
	executor, err := s.store.GetExecutor(exist.ExecutorID)
	if err != nil {
		return nil, err
	}
	if executor.Status != model.ExecutorOnline {
		return nil, model.NewValidationError("executor_id", "执行器不在线，无法启动执行")
	}
	now := time.Now()
	exist.Status = model.ExecutionRunning
	exist.StartedAt = &now
	if err := s.store.UpdateExecution(exist); err != nil {
		return nil, err
	}
	executor.Status = model.ExecutorBusy
	executor.UpdatedAt = now
	if err := s.store.UpdateExecutor(executor); err != nil {
		return nil, err
	}
	s.log.Infof("执行记录 %s 已启动", exist.ID)
	return exist, nil
}

// CompleteExecution 完成执行：running -> completed，记录统计并生成报告。
func (s *Service) CompleteExecution(id string, totalRequests, errorCount int64) (*model.Execution, error) {
	exist, err := s.store.GetExecution(id)
	if err != nil {
		return nil, err
	}
	if !model.ExecutionCanTransition(exist.Status, model.ExecutionCompleted) {
		return nil, model.NewValidationError("status", "执行记录不允许从 "+exist.Status+" 流转到 completed")
	}
	now := time.Now()
	exist.Status = model.ExecutionCompleted
	exist.FinishedAt = &now
	exist.TotalRequests = totalRequests
	exist.ErrorCount = errorCount
	if err := exist.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateExecution(exist); err != nil {
		return nil, err
	}
	s.releaseExecutor(exist.ExecutorID)
	s.log.Infof("执行记录 %s 已完成", exist.ID)
	// 完成时自动聚合指标并生成报告。
	if _, err := s.GenerateReport(exist.ID); err != nil {
		s.log.Warnf("生成报告失败: %v", err)
	}
	return exist, nil
}

// FailExecution 执行失败：running -> failed。
func (s *Service) FailExecution(id string, totalRequests, errorCount int64) (*model.Execution, error) {
	exist, err := s.store.GetExecution(id)
	if err != nil {
		return nil, err
	}
	if !model.ExecutionCanTransition(exist.Status, model.ExecutionFailed) {
		return nil, model.NewValidationError("status", "执行记录不允许从 "+exist.Status+" 流转到 failed")
	}
	now := time.Now()
	exist.Status = model.ExecutionFailed
	exist.FinishedAt = &now
	exist.TotalRequests = totalRequests
	exist.ErrorCount = errorCount
	if err := exist.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateExecution(exist); err != nil {
		return nil, err
	}
	s.releaseExecutor(exist.ExecutorID)
	s.log.Infof("执行记录 %s 已失败", exist.ID)
	return exist, nil
}

// CancelExecution 取消执行：pending/running -> cancelled。
func (s *Service) CancelExecution(id string) (*model.Execution, error) {
	exist, err := s.store.GetExecution(id)
	if err != nil {
		return nil, err
	}
	if !model.ExecutionCanTransition(exist.Status, model.ExecutionCancelled) {
		return nil, model.NewValidationError("status", "执行记录不允许从 "+exist.Status+" 流转到 cancelled")
	}
	now := time.Now()
	exist.Status = model.ExecutionCancelled
	exist.FinishedAt = &now
	if err := s.store.UpdateExecution(exist); err != nil {
		return nil, err
	}
	if exist.ExecutorID != "" {
		s.releaseExecutor(exist.ExecutorID)
	}
	s.log.Infof("执行记录 %s 已取消", exist.ID)
	return exist, nil
}

// releaseExecutor 执行结束后释放执行器资源：busy -> online。
func (s *Service) releaseExecutor(executorID string) {
	executor, err := s.store.GetExecutor(executorID)
	if err != nil {
		return
	}
	if executor.Status != model.ExecutorBusy {
		return
	}
	executor.Status = model.ExecutorOnline
	executor.UpdatedAt = time.Now()
	_ = s.store.UpdateExecutor(executor)
}

// UpdateExecution 更新执行记录统计字段。
func (s *Service) UpdateExecution(id string, input model.Execution) (*model.Execution, error) {
	exist, err := s.store.GetExecution(id)
	if err != nil {
		return nil, err
	}
	if input.TotalRequests >= 0 {
		exist.TotalRequests = input.TotalRequests
	}
	if input.ErrorCount >= 0 {
		exist.ErrorCount = input.ErrorCount
	}
	if input.Status != "" {
		exist.Status = input.Status
	}
	if err := exist.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateExecution(exist); err != nil {
		return nil, err
	}
	return exist, nil
}

// DeleteExecution 删除执行记录及其指标与报告。
func (s *Service) DeleteExecution(id string) error {
	if _, err := s.store.GetExecution(id); err != nil {
		return err
	}
	_ = s.store.DeleteMetricSamplesByExecution(id)
	_ = s.store.DeleteExecution(id)
	return nil
}

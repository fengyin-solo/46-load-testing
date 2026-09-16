package model

import (
	"strings"
	"time"
)

// Execution 状态常量。
const (
	ExecutionPending   = "pending"
	ExecutionRunning   = "running"
	ExecutionCompleted = "completed"
	ExecutionFailed    = "failed"
	ExecutionCancelled = "cancelled"
)

// executionTransitions 执行记录状态机合法流转表。
var executionTransitions = map[string]map[string]bool{
	ExecutionPending:   {ExecutionRunning: true, ExecutionCancelled: true},
	ExecutionRunning:   {ExecutionCompleted: true, ExecutionFailed: true, ExecutionCancelled: true},
	ExecutionCompleted: {},
	ExecutionFailed:    {},
	ExecutionCancelled: {},
}

// ExecutionCanTransition 判断执行记录状态是否可流转。
func ExecutionCanTransition(from, to string) bool {
	if m, ok := executionTransitions[from]; ok {
		return m[to]
	}
	return false
}

// ExecutionIsFinal 判断执行记录是否为终态。
func ExecutionIsFinal(s string) bool {
	return s == ExecutionCompleted || s == ExecutionFailed || s == ExecutionCancelled
}

// Execution 一次压测执行记录。
type Execution struct {
	ID            string     `json:"id"`
	PlanID        string     `json:"plan_id"`
	ExecutorID    string     `json:"executor_id"`
	Status        string     `json:"status"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	TotalRequests int64      `json:"total_requests"`
	ErrorCount    int64      `json:"error_count"`
	CreatedAt     time.Time  `json:"created_at"`
}

// Validate 校验并规范化执行记录字段。
func (e *Execution) Validate() error {
	e.PlanID = strings.TrimSpace(e.PlanID)
	e.ExecutorID = strings.TrimSpace(e.ExecutorID)
	if e.PlanID == "" {
		return NewValidationError("plan_id", "执行记录必须关联测试计划")
	}
	if e.ExecutorID == "" {
		return NewValidationError("executor_id", "执行记录必须关联执行器")
	}
	if e.TotalRequests < 0 {
		return NewValidationError("total_requests", "总请求数不能为负数")
	}
	if e.ErrorCount < 0 {
		return NewValidationError("error_count", "错误数不能为负数")
	}
	if e.ErrorCount > e.TotalRequests {
		return NewValidationError("error_count", "错误数不能超过总请求数")
	}
	if e.Status == "" {
		e.Status = ExecutionPending
	}
	if !IsExecutionStatus(e.Status) {
		return NewValidationError("status", "执行记录状态不合法")
	}
	return nil
}

// IsExecutionStatus 判断是否为合法执行记录状态。
func IsExecutionStatus(s string) bool {
	switch s {
	case ExecutionPending, ExecutionRunning, ExecutionCompleted, ExecutionFailed, ExecutionCancelled:
		return true
	}
	return false
}

// ErrorRate 计算执行记录的错误率，无请求时返回 0。
func (e *Execution) ErrorRate() float64 {
	if e.TotalRequests == 0 {
		return 0
	}
	return float64(e.ErrorCount) / float64(e.TotalRequests)
}

// ExecutionFilter 执行记录列表筛选条件。
type ExecutionFilter struct {
	PlanID     string
	ExecutorID string
	Status     string
}

// Match 判断执行记录是否命中筛选条件。
func (f ExecutionFilter) Match(e *Execution) bool {
	if f.PlanID != "" && e.PlanID != f.PlanID {
		return false
	}
	if f.ExecutorID != "" && e.ExecutorID != f.ExecutorID {
		return false
	}
	if f.Status != "" && e.Status != f.Status {
		return false
	}
	return true
}

package model

import (
	"strings"
	"time"
)

// Schedule 状态常量。
const (
	SchedulePending   = "pending"
	ScheduleTriggered = "triggered"
	ScheduleCancelled = "cancelled"
)

// scheduleTransitions 调度状态机合法流转表。
var scheduleTransitions = map[string]map[string]bool{
	SchedulePending:   {ScheduleTriggered: true, ScheduleCancelled: true},
	ScheduleTriggered: {SchedulePending: true, ScheduleCancelled: true},
	ScheduleCancelled: {SchedulePending: true},
}

// ScheduleCanTransition 判断调度状态是否可流转。
func ScheduleCanTransition(from, to string) bool {
	if m, ok := scheduleTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Schedule 定时调度，按 cron 表达式触发测试计划。
type Schedule struct {
	ID        string     `json:"id"`
	PlanID    string     `json:"plan_id"`
	CronExpr  string     `json:"cron_expr"`
	Status    string     `json:"status"`
	NextRunAt *time.Time `json:"next_run_at"`
	LastRunAt *time.Time `json:"last_run_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Validate 校验并规范化调度字段。
func (s *Schedule) Validate() error {
	s.PlanID = strings.TrimSpace(s.PlanID)
	s.CronExpr = strings.TrimSpace(s.CronExpr)
	if s.PlanID == "" {
		return NewValidationError("plan_id", "调度必须关联测试计划")
	}
	if s.CronExpr == "" {
		return NewValidationError("cron_expr", "调度 cron 表达式不能为空")
	}
	if len(s.CronExpr) > 128 {
		return NewValidationError("cron_expr", "调度 cron 表达式不能超过 128 个字符")
	}
	if !ValidCronExpr(s.CronExpr) {
		return NewValidationError("cron_expr", "调度 cron 表达式不合法，须为 5 段式")
	}
	if s.Status == "" {
		s.Status = SchedulePending
	}
	if !IsScheduleStatus(s.Status) {
		return NewValidationError("status", "调度状态不合法")
	}
	return nil
}

// IsScheduleStatus 判断是否为合法调度状态。
func IsScheduleStatus(s string) bool {
	return s == SchedulePending || s == ScheduleTriggered || s == ScheduleCancelled
}

// ValidCronExpr 校验 cron 表达式是否为 5 段式（分 时 日 月 周）。
func ValidCronExpr(expr string) bool {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return false
	}
	allowed := "*,-/0123456789"
	for _, f := range fields {
		if f == "" {
			return false
		}
		for _, c := range f {
			if !strings.ContainsRune(allowed, c) {
				return false
			}
		}
	}
	return true
}

// ScheduleFilter 调度列表筛选条件。
type ScheduleFilter struct {
	PlanID string
	Status string
}

// Match 判断调度是否命中筛选条件。
func (f ScheduleFilter) Match(s *Schedule) bool {
	if f.PlanID != "" && s.PlanID != f.PlanID {
		return false
	}
	if f.Status != "" && s.Status != f.Status {
		return false
	}
	return true
}

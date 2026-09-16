package model

import (
	"strings"
	"time"
)

// TestPlan 状态常量。
const (
	PlanDraft    = "draft"
	PlanActive   = "active"
	PlanPaused   = "paused"
	PlanArchived = "archived"
)

// planTransitions 测试计划状态机合法流转表。
var planTransitions = map[string]map[string]bool{
	PlanDraft:    {PlanActive: true, PlanArchived: true},
	PlanActive:   {PlanPaused: true, PlanArchived: true},
	PlanPaused:   {PlanActive: true, PlanArchived: true},
	PlanArchived: {},
}

// PlanCanTransition 判断计划状态是否可流转。
func PlanCanTransition(from, to string) bool {
	if m, ok := planTransitions[from]; ok {
		return m[to]
	}
	return false
}

// TestPlan 测试计划。
type TestPlan struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ScenarioID  string    `json:"scenario_id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate 校验并规范化测试计划字段。
func (p *TestPlan) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	p.ScenarioID = strings.TrimSpace(p.ScenarioID)
	p.Description = strings.TrimSpace(p.Description)
	if p.Name == "" {
		return NewValidationError("name", "测试计划名称不能为空")
	}
	if len(p.Name) > 128 {
		return NewValidationError("name", "测试计划名称不能超过 128 个字符")
	}
	if p.ScenarioID == "" {
		return NewValidationError("scenario_id", "测试计划必须关联场景")
	}
	if len(p.Description) > 1024 {
		return NewValidationError("description", "测试计划描述不能超过 1024 个字符")
	}
	if p.Status == "" {
		p.Status = PlanDraft
	}
	if !IsPlanStatus(p.Status) {
		return NewValidationError("status", "测试计划状态不合法")
	}
	return nil
}

// IsPlanStatus 判断是否为合法计划状态。
func IsPlanStatus(s string) bool {
	return s == PlanDraft || s == PlanActive || s == PlanPaused || s == PlanArchived
}

// PlanFilter 测试计划列表筛选条件。
type PlanFilter struct {
	ScenarioID string
	Status     string
	Keyword    string
}

// Match 判断计划是否命中筛选条件。
func (f PlanFilter) Match(p *TestPlan) bool {
	if f.ScenarioID != "" && p.ScenarioID != f.ScenarioID {
		return false
	}
	if f.Status != "" && p.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(p.Name), k) &&
			!strings.Contains(strings.ToLower(p.Description), k) {
			return false
		}
	}
	return true
}

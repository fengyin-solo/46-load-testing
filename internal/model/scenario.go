package model

import (
	"strings"
	"time"
)

// Scenario 状态常量。
const (
	ScenarioDraft    = "draft"
	ScenarioReady    = "ready"
	ScenarioArchived = "archived"
)

// scenarioTransitions 场景状态机合法流转表。
var scenarioTransitions = map[string]map[string]bool{
	ScenarioDraft:    {ScenarioReady: true, ScenarioArchived: true},
	ScenarioReady:    {ScenarioDraft: true, ScenarioArchived: true},
	ScenarioArchived: {ScenarioDraft: true},
}

// ScenarioCanTransition 判断场景状态是否可流转。
func ScenarioCanTransition(from, to string) bool {
	if m, ok := scenarioTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Scenario 压测场景。
type Scenario struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	TargetID    string    `json:"target_id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	Concurrent  int       `json:"concurrent"`
	DurationSec int       `json:"duration_sec"`
	RampUpSec   int       `json:"ramp_up_sec"`
	Body        string    `json:"body"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate 校验并规范化场景字段。
func (s *Scenario) Validate() error {
	s.Name = strings.TrimSpace(s.Name)
	s.TargetID = strings.TrimSpace(s.TargetID)
	s.Method = strings.ToUpper(strings.TrimSpace(s.Method))
	s.Path = strings.TrimSpace(s.Path)
	s.Body = strings.TrimSpace(s.Body)
	if s.Name == "" {
		return NewValidationError("name", "场景名称不能为空")
	}
	if len(s.Name) > 128 {
		return NewValidationError("name", "场景名称不能超过 128 个字符")
	}
	if s.TargetID == "" {
		return NewValidationError("target_id", "场景必须关联目标服务")
	}
	if s.Method == "" {
		s.Method = "GET"
	}
	if !TargetAllowedMethods[s.Method] {
		return NewValidationError("method", "场景 HTTP 方法不合法")
	}
	if s.Path == "" {
		return NewValidationError("path", "场景请求路径不能为空")
	}
	if !strings.HasPrefix(s.Path, "/") {
		return NewValidationError("path", "场景请求路径必须以 / 开头")
	}
	if s.Concurrent <= 0 {
		return NewValidationError("concurrent", "并发数必须大于 0")
	}
	if s.Concurrent > 10000 {
		return NewValidationError("concurrent", "并发数不能超过 10000")
	}
	if s.DurationSec <= 0 {
		return NewValidationError("duration_sec", "压测时长必须大于 0 秒")
	}
	if s.DurationSec > 86400 {
		return NewValidationError("duration_sec", "压测时长不能超过 86400 秒")
	}
	if s.RampUpSec < 0 {
		return NewValidationError("ramp_up_sec", "爬坡时长不能为负数")
	}
	if s.RampUpSec > s.DurationSec {
		return NewValidationError("ramp_up_sec", "爬坡时长不能超过总时长")
	}
	if s.Status == "" {
		s.Status = ScenarioDraft
	}
	if !IsScenarioStatus(s.Status) {
		return NewValidationError("status", "场景状态不合法")
	}
	return nil
}

// IsScenarioStatus 判断是否为合法场景状态。
func IsScenarioStatus(s string) bool {
	return s == ScenarioDraft || s == ScenarioReady || s == ScenarioArchived
}

// ScenarioFilter 场景列表筛选条件。
type ScenarioFilter struct {
	TargetID string
	Method   string
	Status   string
	Keyword  string
}

// Match 判断场景是否命中筛选条件。
func (f ScenarioFilter) Match(s *Scenario) bool {
	if f.TargetID != "" && s.TargetID != f.TargetID {
		return false
	}
	if f.Method != "" && s.Method != strings.ToUpper(f.Method) {
		return false
	}
	if f.Status != "" && s.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(s.Name), k) &&
			!strings.Contains(strings.ToLower(s.Path), k) {
			return false
		}
	}
	return true
}

package model

import (
	"strings"
	"time"
)

// Executor 状态常量。
const (
	ExecutorOnline  = "online"
	ExecutorOffline = "offline"
	ExecutorBusy    = "busy"
)

// executorTransitions 执行器状态机合法流转表。
var executorTransitions = map[string]map[string]bool{
	ExecutorOnline:  {ExecutorBusy: true, ExecutorOffline: true},
	ExecutorBusy:    {ExecutorOnline: true, ExecutorOffline: true},
	ExecutorOffline: {ExecutorOnline: true},
}

// ExecutorCanTransition 判断执行器状态是否可流转。
func ExecutorCanTransition(from, to string) bool {
	if m, ok := executorTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Executor 压测执行器节点。
type Executor struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Addr          string    `json:"addr"`
	Capacity      int       `json:"capacity"`
	Status        string    `json:"status"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Validate 校验并规范化执行器字段。
func (e *Executor) Validate() error {
	e.Name = strings.TrimSpace(e.Name)
	e.Addr = strings.TrimSpace(e.Addr)
	if e.Name == "" {
		return NewValidationError("name", "执行器名称不能为空")
	}
	if len(e.Name) > 128 {
		return NewValidationError("name", "执行器名称不能超过 128 个字符")
	}
	if e.Addr == "" {
		return NewValidationError("addr", "执行器地址不能为空")
	}
	if len(e.Addr) > 256 {
		return NewValidationError("addr", "执行器地址不能超过 256 个字符")
	}
	if e.Capacity <= 0 {
		return NewValidationError("capacity", "执行器容量必须大于 0")
	}
	if e.Capacity > 100000 {
		return NewValidationError("capacity", "执行器容量不能超过 100000")
	}
	if e.Status == "" {
		e.Status = ExecutorOnline
	}
	if !IsExecutorStatus(e.Status) {
		return NewValidationError("status", "执行器状态不合法")
	}
	return nil
}

// IsExecutorStatus 判断是否为合法执行器状态。
func IsExecutorStatus(s string) bool {
	return s == ExecutorOnline || s == ExecutorOffline || s == ExecutorBusy
}

// ExecutorFilter 执行器列表筛选条件。
type ExecutorFilter struct {
	Status  string
	Keyword string
}

// Match 判断执行器是否命中筛选条件。
func (f ExecutorFilter) Match(e *Executor) bool {
	if f.Status != "" && e.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(e.Name), k) &&
			!strings.Contains(strings.ToLower(e.Addr), k) {
			return false
		}
	}
	return true
}

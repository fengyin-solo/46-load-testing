package model

import (
	"strings"
	"time"
)

// Target 状态常量。
const (
	TargetHealthy = "healthy"
	TargetUnhealthy = "unhealthy"
	TargetUnknown  = "unknown"
)

// TargetAllowedMethods 目标服务允许的 HTTP 方法集合。
var TargetAllowedMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"PATCH":   true,
	"DELETE":  true,
	"HEAD":    true,
	"OPTIONS": true,
}

// Target 压测目标服务。
type Target struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Validate 校验并规范化目标服务字段。
func (t *Target) Validate() error {
	t.Name = strings.TrimSpace(t.Name)
	t.URL = strings.TrimSpace(t.URL)
	t.Method = strings.ToUpper(strings.TrimSpace(t.Method))
	if t.Name == "" {
		return NewValidationError("name", "目标服务名称不能为空")
	}
	if len(t.Name) > 128 {
		return NewValidationError("name", "目标服务名称不能超过 128 个字符")
	}
	if t.URL == "" {
		return NewValidationError("url", "目标服务 URL 不能为空")
	}
	if !strings.HasPrefix(t.URL, "http://") && !strings.HasPrefix(t.URL, "https://") {
		return NewValidationError("url", "目标服务 URL 必须以 http:// 或 https:// 开头")
	}
	if t.Method == "" {
		t.Method = "GET"
	}
	if !TargetAllowedMethods[t.Method] {
		return NewValidationError("method", "目标服务 HTTP 方法不合法")
	}
	if t.Headers == nil {
		t.Headers = map[string]string{}
	}
	if t.Status == "" {
		t.Status = TargetUnknown
	}
	if !IsTargetStatus(t.Status) {
		return NewValidationError("status", "目标服务状态不合法")
	}
	return nil
}

// IsTargetStatus 判断是否为合法的目标服务状态。
func IsTargetStatus(s string) bool {
	return s == TargetHealthy || s == TargetUnhealthy || s == TargetUnknown
}

// TargetFilter 目标服务列表筛选条件。
type TargetFilter struct {
	Method  string
	Status  string
	Keyword string
}

// Match 判断目标服务是否命中筛选条件。
func (f TargetFilter) Match(t *Target) bool {
	if f.Method != "" && t.Method != strings.ToUpper(f.Method) {
		return false
	}
	if f.Status != "" && t.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(t.Name), k) &&
			!strings.Contains(strings.ToLower(t.URL), k) {
			return false
		}
	}
	return true
}

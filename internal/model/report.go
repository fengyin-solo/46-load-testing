package model

import (
	"strings"
	"time"
)

// Report 执行完成后的压测报告。
type Report struct {
	ID          string    `json:"id"`
	ExecutionID string    `json:"execution_id"`
	Summary     string    `json:"summary"`
	Conclusion  string    `json:"conclusion"`
	GeneratedAt time.Time `json:"generated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// Validate 校验并规范化报告字段。
func (r *Report) Validate() error {
	r.ExecutionID = strings.TrimSpace(r.ExecutionID)
	r.Summary = strings.TrimSpace(r.Summary)
	r.Conclusion = strings.TrimSpace(r.Conclusion)
	if r.ExecutionID == "" {
		return NewValidationError("execution_id", "报告必须关联执行记录")
	}
	if len(r.Summary) > 8192 {
		return NewValidationError("summary", "报告摘要不能超过 8192 个字符")
	}
	if len(r.Conclusion) > 2048 {
		return NewValidationError("conclusion", "报告结论不能超过 2048 个字符")
	}
	return nil
}

// ReportFilter 报告列表筛选条件。
type ReportFilter struct {
	ExecutionID string
	Keyword     string
}

// Match 判断报告是否命中筛选条件。
func (f ReportFilter) Match(r *Report) bool {
	if f.ExecutionID != "" && r.ExecutionID != f.ExecutionID {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(r.Summary), k) &&
			!strings.Contains(strings.ToLower(r.Conclusion), k) {
			return false
		}
	}
	return true
}

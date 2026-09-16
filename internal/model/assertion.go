package model

import (
	"strings"
	"time"
)

// Assertion 指标维度常量。
const (
	MetricTPS        = "tps"
	MetricAvgLatency = "avg_latency"
	MetricP99        = "p99"
	MetricErrorRate  = "error_rate"
)

// Assertion 运算符常量。
const (
	OperatorLT = "lt"
	OperatorGT = "gt"
	OperatorEQ = "eq"
)

// Assertion 结果常量。
const (
	AssertionPass = "pass"
	AssertionFail = "fail"
)

// IsValidMetric 判断是否为合法断言指标。
func IsValidMetric(m string) bool {
	switch m {
	case MetricTPS, MetricAvgLatency, MetricP99, MetricErrorRate:
		return true
	}
	return false
}

// IsValidOperator 判断是否为合法断言运算符。
func IsValidOperator(o string) bool {
	switch o {
	case OperatorLT, OperatorGT, OperatorEQ:
		return true
	}
	return false
}

// Assertion 断言，针对某次执行的指标判定通过/失败。
type Assertion struct {
	ID          string    `json:"id"`
	ExecutionID string    `json:"execution_id"`
	Metric      string    `json:"metric"`
	Operator    string    `json:"operator"`
	Threshold   float64   `json:"threshold"`
	Result      string    `json:"result"`
	CreatedAt   time.Time `json:"created_at"`
}

// Validate 校验并规范化断言字段。
func (a *Assertion) Validate() error {
	a.ExecutionID = strings.TrimSpace(a.ExecutionID)
	a.Metric = strings.TrimSpace(a.Metric)
	a.Operator = strings.TrimSpace(a.Operator)
	if a.ExecutionID == "" {
		return NewValidationError("execution_id", "断言必须关联执行记录")
	}
	if !IsValidMetric(a.Metric) {
		return NewValidationError("metric", "断言指标不合法")
	}
	if !IsValidOperator(a.Operator) {
		return NewValidationError("operator", "断言运算符不合法")
	}
	if a.Threshold < 0 {
		return NewValidationError("threshold", "断言阈值不能为负数")
	}
	if a.Result == "" {
		a.Result = AssertionPass
	}
	if !IsValidAssertionResult(a.Result) {
		return NewValidationError("result", "断言结果不合法")
	}
	return nil
}

// IsValidAssertionResult 判断是否为合法断言结果。
func IsValidAssertionResult(r string) bool {
	return r == AssertionPass || r == AssertionFail
}

// Evaluate 根据实际指标值判定断言结果。
func (a *Assertion) Evaluate(actual float64) bool {
	switch a.Operator {
	case OperatorLT:
		return actual < a.Threshold
	case OperatorGT:
		return actual > a.Threshold
	case OperatorEQ:
		const epsilon = 1e-9
		diff := actual - a.Threshold
		return diff > -epsilon && diff < epsilon
	default:
		return false
	}
}

// AssertionFilter 断言列表筛选条件。
type AssertionFilter struct {
	ExecutionID string
	Metric      string
	Result      string
}

// Match 判断断言是否命中筛选条件。
func (f AssertionFilter) Match(a *Assertion) bool {
	if f.ExecutionID != "" && a.ExecutionID != f.ExecutionID {
		return false
	}
	if f.Metric != "" && a.Metric != f.Metric {
		return false
	}
	if f.Result != "" && a.Result != f.Result {
		return false
	}
	return true
}

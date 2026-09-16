package model

import (
	"strings"
	"time"
)

// MetricSample 指标样本，记录某执行在某一时刻的性能指标。
type MetricSample struct {
	ID           string    `json:"id"`
	ExecutionID  string    `json:"execution_id"`
	Timestamp    time.Time `json:"timestamp"`
	TPS          float64   `json:"tps"`
	AvgLatencyMs float64   `json:"avg_latency_ms"`
	P99LatencyMs float64   `json:"p99_latency_ms"`
	ErrorRate    float64   `json:"error_rate"`
	CreatedAt    time.Time `json:"created_at"`
}

// Validate 校验并规范化指标样本字段。
func (m *MetricSample) Validate() error {
	m.ExecutionID = strings.TrimSpace(m.ExecutionID)
	if m.ExecutionID == "" {
		return NewValidationError("execution_id", "指标样本必须关联执行记录")
	}
	if m.TPS < 0 {
		return NewValidationError("tps", "TPS 不能为负数")
	}
	if m.AvgLatencyMs < 0 {
		return NewValidationError("avg_latency_ms", "平均延迟不能为负数")
	}
	if m.P99LatencyMs < 0 {
		return NewValidationError("p99_latency_ms", "P99 延迟不能为负数")
	}
	if m.P99LatencyMs < m.AvgLatencyMs {
		return NewValidationError("p99_latency_ms", "P99 延迟不能小于平均延迟")
	}
	if m.ErrorRate < 0 || m.ErrorRate > 1 {
		return NewValidationError("error_rate", "错误率必须在 0 到 1 之间")
	}
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}
	return nil
}

// MetricSampleFilter 指标样本列表筛选条件。
type MetricSampleFilter struct {
	ExecutionID string
	MinTPS      float64
	MaxErrorRate float64
}

// Match 判断指标样本是否命中筛选条件。
func (f MetricSampleFilter) Match(m *MetricSample) bool {
	if f.ExecutionID != "" && m.ExecutionID != f.ExecutionID {
		return false
	}
	if f.MinTPS > 0 && m.TPS < f.MinTPS {
		return false
	}
	if f.MaxErrorRate > 0 && m.ErrorRate > f.MaxErrorRate {
		return false
	}
	return true
}

// MetricSamples 指标样本切片，用于聚合计算。
type MetricSamples []*MetricSample

// AvgTPS 计算平均 TPS。
func (ms MetricSamples) AvgTPS() float64 {
	if len(ms) == 0 {
		return 0
	}
	var sum float64
	for _, m := range ms {
		sum += m.TPS
	}
	return sum / float64(len(ms))
}

// AvgLatency 计算平均延迟。
func (ms MetricSamples) AvgLatency() float64 {
	if len(ms) == 0 {
		return 0
	}
	var sum float64
	for _, m := range ms {
		sum += m.AvgLatencyMs
	}
	return sum / float64(len(ms))
}

// MaxP99 计算 P99 延迟峰值。
func (ms MetricSamples) MaxP99() float64 {
	var max float64
	for _, m := range ms {
		if m.P99LatencyMs > max {
			max = m.P99LatencyMs
		}
	}
	return max
}

// AvgErrorRate 计算平均错误率。
func (ms MetricSamples) AvgErrorRate() float64 {
	if len(ms) == 0 {
		return 0
	}
	var sum float64
	for _, m := range ms {
		sum += m.ErrorRate
	}
	return sum / float64(len(ms))
}

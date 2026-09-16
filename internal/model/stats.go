package model

import "time"

// MetricAggregation 指标样本聚合结果。
type MetricAggregation struct {
	ExecutionID  string  `json:"execution_id"`
	SampleCount  int     `json:"sample_count"`
	AvgTPS       float64 `json:"avg_tps"`
	AvgLatency   float64 `json:"avg_latency_ms"`
	MaxP99       float64 `json:"max_p99_ms"`
	AvgErrorRate float64 `json:"avg_error_rate"`
}

// ExecutionAggregation 执行记录维度聚合。
type ExecutionAggregation struct {
	ExecutionID   string  `json:"execution_id"`
	PlanID        string  `json:"plan_id"`
	Status        string  `json:"status"`
	TotalRequests int64   `json:"total_requests"`
	ErrorCount    int64   `json:"error_count"`
	ErrorRate     float64 `json:"error_rate"`
	AvgTPS        float64 `json:"avg_tps"`
	AvgLatency    float64 `json:"avg_latency_ms"`
	MaxP99        float64 `json:"max_p99_ms"`
	SampleCount   int     `json:"sample_count"`
}

// PlanAggregation 测试计划维度聚合。
type PlanAggregation struct {
	PlanID          string  `json:"plan_id"`
	PlanName        string  `json:"plan_name"`
	ExecutionCount  int     `json:"execution_count"`
	TotalRequests   int64   `json:"total_requests"`
	ErrorCount      int64   `json:"error_count"`
	ErrorRate       float64 `json:"error_rate"`
	AvgTPS          float64 `json:"avg_tps"`
	AvgLatency      float64 `json:"avg_latency_ms"`
	MaxP99          float64 `json:"max_p99_ms"`
}

// ExecutorAggregation 执行器维度聚合。
type ExecutorAggregation struct {
	ExecutorID      string  `json:"executor_id"`
	ExecutorName    string  `json:"executor_name"`
	ExecutionCount  int     `json:"execution_count"`
	TotalRequests   int64   `json:"total_requests"`
	ErrorCount      int64   `json:"error_count"`
	ErrorRate       float64 `json:"error_rate"`
}

// TopScenario 排名场景。
type TopScenario struct {
	ScenarioID     string  `json:"scenario_id"`
	ScenarioName   string  `json:"scenario_name"`
	ExecutionCount int     `json:"execution_count"`
	TotalRequests  int64   `json:"total_requests"`
	AvgTPS         float64 `json:"avg_tps"`
}

// OverviewStats 全局统计快照。
type OverviewStats struct {
	TargetCount       int                  `json:"target_count"`
	ScenarioCount     int                  `json:"scenario_count"`
	PlanCount         int                  `json:"plan_count"`
	ExecutorCount     int                  `json:"executor_count"`
	ExecutionCount    int                  `json:"execution_count"`
	MetricCount       int                  `json:"metric_count"`
	ReportCount       int                  `json:"report_count"`
	ScheduleCount     int                  `json:"schedule_count"`
	RunningExecutions int                  `json:"running_executions"`
	TotalRequests     int64                `json:"total_requests"`
	TotalErrors       int64                `json:"total_errors"`
	GlobalErrorRate   float64              `json:"global_error_rate"`
	GlobalAvgTPS      float64              `json:"global_avg_tps"`
	GlobalAvgLatency  float64              `json:"global_avg_latency_ms"`
	GlobalMaxP99      float64              `json:"global_max_p99_ms"`
	ByStatus          map[string]int       `json:"by_status"`
	GeneratedAt       time.Time            `json:"generated_at"`
}

// ExportSnapshot 导出汇总快照。
type ExportSnapshot struct {
	GeneratedAt time.Time               `json:"generated_at"`
	Overview    OverviewStats           `json:"overview"`
	Plans       []*PlanAggregation      `json:"plans"`
	Executors   []*ExecutorAggregation  `json:"executors"`
	TopScenarios []*TopScenario         `json:"top_scenarios"`
}

// Package store 定义数据访问接口与内存实现。
package store

import (
	"errors"

	"loadtest/internal/model"
)

var (
	// ErrNotFound 表示记录不存在。
	ErrNotFound = errors.New("记录不存在")
	// ErrConflict 表示记录已存在或状态冲突。
	ErrConflict = errors.New("记录已存在或状态冲突")
)

// Store 聚合全部实体的数据访问方法，便于测试时替换实现。
type Store interface {
	// Target 目标服务
	CreateTarget(t *model.Target) error
	GetTarget(id string) (*model.Target, error)
	GetTargetByURLMethod(url, method string) (*model.Target, error)
	ListTargets() []*model.Target
	UpdateTarget(t *model.Target) error
	DeleteTarget(id string) error

	// Scenario 场景
	CreateScenario(s *model.Scenario) error
	GetScenario(id string) (*model.Scenario, error)
	GetScenarioByName(name string) (*model.Scenario, error)
	ListScenarios() []*model.Scenario
	UpdateScenario(s *model.Scenario) error
	DeleteScenario(id string) error

	// TestPlan 测试计划
	CreateTestPlan(p *model.TestPlan) error
	GetTestPlan(id string) (*model.TestPlan, error)
	ListTestPlans() []*model.TestPlan
	UpdateTestPlan(p *model.TestPlan) error
	DeleteTestPlan(id string) error

	// Executor 执行器
	CreateExecutor(e *model.Executor) error
	GetExecutor(id string) (*model.Executor, error)
	ListExecutors() []*model.Executor
	UpdateExecutor(e *model.Executor) error
	DeleteExecutor(id string) error

	// Execution 执行记录
	CreateExecution(e *model.Execution) error
	GetExecution(id string) (*model.Execution, error)
	ListExecutions() []*model.Execution
	UpdateExecution(e *model.Execution) error
	DeleteExecution(id string) error

	// MetricSample 指标样本
	CreateMetricSample(m *model.MetricSample) error
	CreateMetricSamples(ms []*model.MetricSample) error
	GetMetricSample(id string) (*model.MetricSample, error)
	ListMetricSamples() []*model.MetricSample
	DeleteMetricSample(id string) error
	DeleteMetricSamplesByExecution(executionID string) error

	// Assertion 断言
	CreateAssertion(a *model.Assertion) error
	GetAssertion(id string) (*model.Assertion, error)
	ListAssertions() []*model.Assertion
	DeleteAssertion(id string) error

	// Report 报告
	CreateReport(r *model.Report) error
	GetReport(id string) (*model.Report, error)
	GetReportByExecution(executionID string) (*model.Report, error)
	ListReports() []*model.Report
	DeleteReport(id string) error

	// Schedule 调度
	CreateSchedule(s *model.Schedule) error
	GetSchedule(id string) (*model.Schedule, error)
	ListSchedules() []*model.Schedule
	UpdateSchedule(s *model.Schedule) error
	DeleteSchedule(id string) error
}

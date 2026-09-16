package store

import (
	"sync"

	"loadtest/internal/model"
)

// MemoryStore 基于内存的 Store 实现，使用互斥锁保证并发安全。
type MemoryStore struct {
	mu       sync.RWMutex
	targets  map[string]*model.Target
	scenarios map[string]*model.Scenario
	testplans map[string]*model.TestPlan
	executors map[string]*model.Executor
	executions map[string]*model.Execution
	metrics   map[string]*model.MetricSample
	assertions map[string]*model.Assertion
	reports   map[string]*model.Report
	schedules map[string]*model.Schedule
}

// NewMemoryStore 创建空的 MemoryStore。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		targets:    make(map[string]*model.Target),
		scenarios:  make(map[string]*model.Scenario),
		testplans:  make(map[string]*model.TestPlan),
		executors:  make(map[string]*model.Executor),
		executions: make(map[string]*model.Execution),
		metrics:    make(map[string]*model.MetricSample),
		assertions: make(map[string]*model.Assertion),
		reports:    make(map[string]*model.Report),
		schedules:  make(map[string]*model.Schedule),
	}
}

var _ Store = (*MemoryStore)(nil)

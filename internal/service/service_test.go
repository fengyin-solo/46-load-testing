package service

import (
	"testing"

	"loadtest/internal/config"
	"loadtest/internal/model"
	"loadtest/internal/store"
	"loadtest/pkg/logger"
)

func newTestService() *Service {
	cfg := &config.Config{MaxPageSize: 100, APIKey: "test-key", RateLimit: 1000, RateWindowSec: 60}
	log := logger.NewLevel(logger.LevelError)
	return New(store.NewMemoryStore(), log, cfg)
}

// setupChain 构造 target -> scenario -> plan -> executor 完整链路。
func setupChain(t *testing.T, s *Service) (targetID, scenarioID, planID, executorID string) {
	t.Helper()
	target, err := s.CreateTarget(model.Target{Name: "目标服务", URL: "https://api.example.com", Method: "GET"})
	if err != nil {
		t.Fatalf("创建目标失败: %v", err)
	}
	scenario, err := s.CreateScenario(model.Scenario{
		Name: "首页压测", TargetID: target.ID, Method: "GET", Path: "/",
		Concurrent: 100, DurationSec: 60, RampUpSec: 5,
	})
	if err != nil {
		t.Fatalf("创建场景失败: %v", err)
	}
	plan, err := s.CreateTestPlan(model.TestPlan{Name: "计划A", ScenarioID: scenario.ID})
	if err != nil {
		t.Fatalf("创建计划失败: %v", err)
	}
	executor, err := s.CreateExecutor(model.Executor{Name: "worker-1", Addr: "10.0.0.1:9000", Capacity: 500})
	if err != nil {
		t.Fatalf("创建执行器失败: %v", err)
	}
	return target.ID, scenario.ID, plan.ID, executor.ID
}

func TestCreateScenarioCrossEntityValidation(t *testing.T) {
	s := newTestService()
	// 目标服务不存在时应报错
	_, err := s.CreateScenario(model.Scenario{
		Name: "非法场景", TargetID: "missing", Method: "GET", Path: "/",
		Concurrent: 10, DurationSec: 60,
	})
	if err == nil {
		t.Fatal("目标服务不存在应报错")
	}
	if !model.IsValidationError(err) {
		t.Fatalf("应为校验错误, 实际 %v", err)
	}
}

func TestCreateExecutionCrossEntityValidation(t *testing.T) {
	s := newTestService()
	_, _, planID, _ := setupChain(t, s)
	// 执行器不存在应报错
	_, err := s.CreateExecution(model.Execution{PlanID: planID, ExecutorID: "missing"})
	if err == nil {
		t.Fatal("执行器不存在应报错")
	}
}

func TestExecutionStateMachine(t *testing.T) {
	s := newTestService()
	_, _, planID, executorID := setupChain(t, s)
	ex, err := s.CreateExecution(model.Execution{PlanID: planID, ExecutorID: executorID})
	if err != nil {
		t.Fatalf("创建执行失败: %v", err)
	}
	if ex.Status != model.ExecutionPending {
		t.Fatalf("初始状态应为 pending, 实际 %s", ex.Status)
	}
	// pending -> running
	ex, err = s.StartExecution(ex.ID)
	if err != nil {
		t.Fatalf("启动失败: %v", err)
	}
	if ex.Status != model.ExecutionRunning {
		t.Fatalf("状态应为 running")
	}
	// 执行器应变为 busy
	e, _ := s.store.GetExecutor(executorID)
	if e.Status != model.ExecutorBusy {
		t.Fatalf("执行器应变为 busy, 实际 %s", e.Status)
	}
	// running -> completed
	ex, err = s.CompleteExecution(ex.ID, 10000, 100)
	if err != nil {
		t.Fatalf("完成失败: %v", err)
	}
	if ex.Status != model.ExecutionCompleted {
		t.Fatalf("状态应为 completed")
	}
	// 执行器应释放为 online
	e, _ = s.store.GetExecutor(executorID)
	if e.Status != model.ExecutorOnline {
		t.Fatalf("执行器应释放为 online, 实际 %s", e.Status)
	}
	// 完成后不应再流转
	if _, err := s.StartExecution(ex.ID); err == nil {
		t.Fatal("completed 状态不应再启动")
	}
}

func TestInvalidExecutionTransition(t *testing.T) {
	s := newTestService()
	_, _, planID, executorID := setupChain(t, s)
	ex, _ := s.CreateExecution(model.Execution{PlanID: planID, ExecutorID: executorID})
	// pending 不能直接 completed
	if _, err := s.CompleteExecution(ex.ID, 1, 0); err == nil {
		t.Fatal("pending 不能直接 completed")
	}
}

func TestTestPlanStateMachine(t *testing.T) {
	s := newTestService()
	_, _, planID, _ := setupChain(t, s)
	p, err := s.TransitionPlan(planID, model.PlanActive)
	if err != nil {
		t.Fatalf("draft->active 应成功: %v", err)
	}
	if p.Status != model.PlanActive {
		t.Fatalf("状态应为 active")
	}
	// active -> paused
	if _, err := s.TransitionPlan(planID, model.PlanPaused); err != nil {
		t.Fatalf("active->paused 应成功: %v", err)
	}
	// paused -> active
	if _, err := s.TransitionPlan(planID, model.PlanActive); err != nil {
		t.Fatalf("paused->active 应成功: %v", err)
	}
	// active -> archived
	if _, err := s.TransitionPlan(planID, model.PlanArchived); err != nil {
		t.Fatalf("active->archived 应成功: %v", err)
	}
	// archived 不能再流转
	if _, err := s.TransitionPlan(planID, model.PlanActive); err == nil {
		t.Fatal("archived 不应再流转")
	}
}

func TestScenarioStateMachine(t *testing.T) {
	s := newTestService()
	_, scenarioID, _, _ := setupChain(t, s)
	if _, err := s.TransitionScenario(scenarioID, model.ScenarioReady); err != nil {
		t.Fatalf("draft->ready 应成功: %v", err)
	}
	if _, err := s.TransitionScenario(scenarioID, model.ScenarioArchived); err != nil {
		t.Fatalf("ready->archived 应成功: %v", err)
	}
	// archived -> ready 非法
	if _, err := s.TransitionScenario(scenarioID, model.ScenarioReady); err == nil {
		t.Fatal("archived->ready 应非法")
	}
}

func TestExecutorStateMachine(t *testing.T) {
	s := newTestService()
	_, _, _, executorID := setupChain(t, s)
	if _, err := s.TransitionExecutor(executorID, model.ExecutorBusy); err != nil {
		t.Fatalf("online->busy 应成功: %v", err)
	}
	if _, err := s.TransitionExecutor(executorID, model.ExecutorOnline); err != nil {
		t.Fatalf("busy->online 应成功: %v", err)
	}
	if _, err := s.TransitionExecutor(executorID, model.ExecutorOffline); err != nil {
		t.Fatalf("online->offline 应成功: %v", err)
	}
	if _, err := s.TransitionExecutor(executorID, model.ExecutorOnline); err != nil {
		t.Fatalf("offline->online 应成功: %v", err)
	}
}

func TestAssertionEvaluation(t *testing.T) {
	s := newTestService()
	_, _, planID, executorID := setupChain(t, s)
	ex, _ := s.CreateExecution(model.Execution{PlanID: planID, ExecutorID: executorID})
	// 写入指标样本：TPS 平均 120
	_, err := s.CreateMetricSamples(ex.ID, []model.MetricSample{
		{TPS: 100, AvgLatencyMs: 50, P99LatencyMs: 80, ErrorRate: 0.01},
		{TPS: 140, AvgLatencyMs: 60, P99LatencyMs: 90, ErrorRate: 0.02},
	})
	if err != nil {
		t.Fatalf("写入样本失败: %v", err)
	}
	// TPS > 100 应通过
	a, err := s.CreateAssertion(model.Assertion{
		ExecutionID: ex.ID, Metric: model.MetricTPS, Operator: model.OperatorGT, Threshold: 100,
	})
	if err != nil {
		t.Fatalf("创建断言失败: %v", err)
	}
	if a.Result != model.AssertionPass {
		t.Fatalf("断言应通过, 实际 %s", a.Result)
	}
	// TPS < 200 应通过（平均 120 < 200）
	a2, err := s.CreateAssertion(model.Assertion{
		ExecutionID: ex.ID, Metric: model.MetricTPS, Operator: model.OperatorLT, Threshold: 200,
	})
	if err != nil {
		t.Fatalf("创建断言失败: %v", err)
	}
	if a2.Result != model.AssertionPass {
		t.Fatalf("断言应通过, 实际 %s", a2.Result)
	}
	// TPS > 200 应失败（平均 120 < 200）
	a3, err := s.CreateAssertion(model.Assertion{
		ExecutionID: ex.ID, Metric: model.MetricTPS, Operator: model.OperatorGT, Threshold: 200,
	})
	if err != nil {
		t.Fatalf("创建断言失败: %v", err)
	}
	if a3.Result != model.AssertionFail {
		t.Fatalf("断言应失败, 实际 %s", a3.Result)
	}
}

func TestStatsAggregation(t *testing.T) {
	s := newTestService()
	_, _, planID, executorID := setupChain(t, s)
	ex, _ := s.CreateExecution(model.Execution{PlanID: planID, ExecutorID: executorID})
	if _, err := s.CreateMetricSamples(ex.ID, []model.MetricSample{
		{TPS: 100, AvgLatencyMs: 50, P99LatencyMs: 80, ErrorRate: 0.01},
		{TPS: 200, AvgLatencyMs: 70, P99LatencyMs: 120, ErrorRate: 0.03},
	}); err != nil {
		t.Fatalf("写入样本失败: %v", err)
	}
	overview, err := s.Overview()
	if err != nil {
		t.Fatalf("概览失败: %v", err)
	}
	if overview.TargetCount != 1 || overview.ScenarioCount != 1 || overview.PlanCount != 1 {
		t.Fatalf("概览计数错误: %+v", overview)
	}
	// 平均 TPS = (100+200)/2 = 150
	if overview.GlobalAvgTPS != 150 {
		t.Fatalf("平均 TPS 应为 150, 实际 %f", overview.GlobalAvgTPS)
	}
	// 按计划聚合
	plans, err := s.AggregationsByPlan()
	if err != nil {
		t.Fatalf("按计划聚合失败: %v", err)
	}
	if len(plans) != 1 || plans[0].PlanID != planID {
		t.Fatalf("计划聚合错误: %+v", plans)
	}
	// TOP 场景
	top, err := s.TopScenarios(10)
	if err != nil {
		t.Fatalf("TOP 场景失败: %v", err)
	}
	if len(top) != 1 {
		t.Fatalf("TOP 场景应 1 条, 实际 %d", len(top))
	}
	// 导出快照
	snap, err := s.ExportSnapshot(10)
	if err != nil {
		t.Fatalf("导出快照失败: %v", err)
	}
	if snap.Overview.TargetCount != 1 {
		t.Fatalf("快照概览错误")
	}
}

func TestReportGeneration(t *testing.T) {
	s := newTestService()
	_, _, planID, executorID := setupChain(t, s)
	ex, _ := s.CreateExecution(model.Execution{PlanID: planID, ExecutorID: executorID})
	// 未完成时不应生成报告
	if _, err := s.GenerateReport(ex.ID); err == nil {
		t.Fatal("未完成不应生成报告")
	}
	// 启动后完成执行，自动生成报告
	if _, err := s.StartExecution(ex.ID); err != nil {
		t.Fatalf("启动失败: %v", err)
	}
	if _, err := s.CompleteExecution(ex.ID, 1000, 10); err != nil {
		t.Fatalf("完成失败: %v", err)
	}
	report, err := s.GetReportByExecution(ex.ID)
	if err != nil {
		t.Fatalf("应自动生成报告: %v", err)
	}
	if report.ExecutionID != ex.ID {
		t.Fatalf("报告执行 ID 错误")
	}
	if report.Summary == "" || report.Conclusion == "" {
		t.Fatalf("报告摘要与结论不能为空")
	}
}

func TestListPaginationAndFilter(t *testing.T) {
	s := newTestService()
	t1, _ := s.CreateTarget(model.Target{Name: "A服务", URL: "https://a.com", Method: "GET"})
	_, _ = s.CreateTarget(model.Target{Name: "B服务", URL: "https://b.com", Method: "POST"})
	// 按 method 筛选
	items, total, err := s.ListTargets(model.TargetFilter{Method: "GET"}, 1, 10)
	if err != nil {
		t.Fatalf("列表失败: %v", err)
	}
	if total != 1 || items[0].ID != t1.ID {
		t.Fatalf("筛选结果错误: %d %+v", total, items)
	}
	// 分页：size=1 page=2
	items, total, err = s.ListTargets(model.TargetFilter{}, 2, 1)
	if err != nil {
		t.Fatalf("列表失败: %v", err)
	}
	if total != 2 || len(items) != 1 {
		t.Fatalf("分页结果错误: %d %d", total, len(items))
	}
}

func TestBatchTransitionPlans(t *testing.T) {
	s := newTestService()
	_, scenarioID, _, _ := setupChain(t, s)
	p1, _ := s.CreateTestPlan(model.TestPlan{Name: "P1", ScenarioID: scenarioID})
	p2, _ := s.CreateTestPlan(model.TestPlan{Name: "P2", ScenarioID: scenarioID})
	updated, err := s.BatchTransitionPlans([]string{p1.ID, p2.ID}, model.PlanActive)
	if err != nil {
		t.Fatalf("批量流转失败: %v", err)
	}
	if len(updated) != 2 {
		t.Fatalf("应更新 2 个计划")
	}
	for _, p := range updated {
		if p.Status != model.PlanActive {
			t.Fatalf("状态应为 active")
		}
	}
}

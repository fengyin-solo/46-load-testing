package store

import (
	"testing"
	"time"

	"loadtest/internal/model"
)

func newTestTarget() *model.Target {
	return &model.Target{
		ID: "t1", Name: "用户服务", URL: "https://api.example.com/users",
		Method: "GET", Status: model.TargetUnknown,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}

func TestTargetCRUD(t *testing.T) {
	s := NewMemoryStore()
	// Create
	if err := s.CreateTarget(newTestTarget()); err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	// Get
	got, err := s.GetTarget("t1")
	if err != nil || got.Name != "用户服务" {
		t.Fatalf("查询失败: %v %v", got, err)
	}
	// 冲突：同 URL+Method
	if err := s.CreateTarget(&model.Target{ID: "t2", URL: "https://api.example.com/users", Method: "GET"}); err != ErrConflict {
		t.Fatalf("应返回冲突, 实际 %v", err)
	}
	// GetByURLMethod
	got2, err := s.GetTargetByURLMethod("https://api.example.com/users", "GET")
	if err != nil || got2.ID != "t1" {
		t.Fatalf("按 URL+Method 查询失败: %v", err)
	}
	// Update
	got.Name = "用户服务V2"
	if err := s.UpdateTarget(got); err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	// List
	if list := s.ListTargets(); len(list) != 1 {
		t.Fatalf("列表长度应为 1, 实际 %d", len(list))
	}
	// Delete
	if err := s.DeleteTarget("t1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	// 不存在
	if _, err := s.GetTarget("t1"); err != ErrNotFound {
		t.Fatalf("应返回不存在, 实际 %v", err)
	}
	if err := s.DeleteTarget("t1"); err != ErrNotFound {
		t.Fatalf("删除不存在记录应报错, 实际 %v", err)
	}
}

func TestScenarioCRUDAndConflict(t *testing.T) {
	s := NewMemoryStore()
	sc := &model.Scenario{ID: "s1", Name: "首页压测", TargetID: "t1"}
	if err := s.CreateScenario(sc); err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if err := s.CreateScenario(&model.Scenario{ID: "s2", Name: "首页压测"}); err != ErrConflict {
		t.Fatalf("名称冲突应报错, 实际 %v", err)
	}
	if _, err := s.GetScenarioByName("首页压测"); err != nil {
		t.Fatalf("按名称查询失败: %v", err)
	}
	if _, err := s.GetScenario("s1"); err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if err := s.DeleteScenario("s1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
}

func TestTestPlanCRUD(t *testing.T) {
	s := NewMemoryStore()
	p := &model.TestPlan{ID: "p1", Name: "大促压测计划", ScenarioID: "s1"}
	if err := s.CreateTestPlan(p); err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if _, err := s.GetTestPlan("p1"); err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	p.Status = model.PlanActive
	if err := s.UpdateTestPlan(p); err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	if len(s.ListTestPlans()) != 1 {
		t.Fatalf("列表长度错误")
	}
	if err := s.DeleteTestPlan("p1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
}

func TestExecutorCRUD(t *testing.T) {
	s := NewMemoryStore()
	e := &model.Executor{ID: "e1", Name: "worker-1", Addr: "10.0.0.1:9000", Capacity: 100}
	if err := s.CreateExecutor(e); err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if err := s.CreateExecutor(&model.Executor{ID: "e2", Name: "worker-1"}); err != ErrConflict {
		t.Fatalf("名称冲突应报错")
	}
	if _, err := s.GetExecutor("e1"); err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if err := s.DeleteExecutor("e1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
}

func TestExecutionCRUD(t *testing.T) {
	s := NewMemoryStore()
	ex := &model.Execution{ID: "x1", PlanID: "p1", ExecutorID: "e1", Status: model.ExecutionPending}
	if err := s.CreateExecution(ex); err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if _, err := s.GetExecution("x1"); err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	ex.Status = model.ExecutionRunning
	if err := s.UpdateExecution(ex); err != nil {
		t.Fatalf("更新失败: %v", err)
	}
	if err := s.DeleteExecution("x1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
}

func TestMetricSampleCRUDAndBatch(t *testing.T) {
	s := NewMemoryStore()
	m := &model.MetricSample{ID: "m1", ExecutionID: "x1", TPS: 100.5}
	if err := s.CreateMetricSample(m); err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	batch := []*model.MetricSample{
		{ID: "m2", ExecutionID: "x1", TPS: 120},
		{ID: "m3", ExecutionID: "x1", TPS: 130},
	}
	if err := s.CreateMetricSamples(batch); err != nil {
		t.Fatalf("批量创建失败: %v", err)
	}
	if len(s.ListMetricSamples()) != 3 {
		t.Fatalf("应有 3 条样本, 实际 %d", len(s.ListMetricSamples()))
	}
	if err := s.DeleteMetricSample("m1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
	if err := s.DeleteMetricSamplesByExecution("x1"); err != nil {
		t.Fatalf("按执行删除失败: %v", err)
	}
	if len(s.ListMetricSamples()) != 0 {
		t.Fatalf("应清空样本")
	}
}

func TestAssertionCRUD(t *testing.T) {
	s := NewMemoryStore()
	a := &model.Assertion{ID: "a1", ExecutionID: "x1", Metric: model.MetricTPS, Operator: model.OperatorGT, Threshold: 100}
	if err := s.CreateAssertion(a); err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if _, err := s.GetAssertion("a1"); err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if err := s.DeleteAssertion("a1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
}

func TestReportCRUDAndConflict(t *testing.T) {
	s := NewMemoryStore()
	r := &model.Report{ID: "r1", ExecutionID: "x1"}
	if err := s.CreateReport(r); err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if err := s.CreateReport(&model.Report{ID: "r2", ExecutionID: "x1"}); err != ErrConflict {
		t.Fatalf("执行记录唯一性应冲突")
	}
	if _, err := s.GetReportByExecution("x1"); err != nil {
		t.Fatalf("按执行查询失败: %v", err)
	}
	if err := s.DeleteReport("r1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
}

func TestScheduleCRUD(t *testing.T) {
	s := NewMemoryStore()
	sc := &model.Schedule{ID: "c1", PlanID: "p1", CronExpr: "0 * * * *"}
	if err := s.CreateSchedule(sc); err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if _, err := s.GetSchedule("c1"); err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if err := s.DeleteSchedule("c1"); err != nil {
		t.Fatalf("删除失败: %v", err)
	}
}

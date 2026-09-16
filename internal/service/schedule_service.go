package service

import (
	"sort"
	"time"

	"loadtest/internal/model"
	"loadtest/pkg/idgen"
)

// CreateSchedule 创建调度，校验测试计划存在。
func (s *Service) CreateSchedule(input model.Schedule) (*model.Schedule, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetTestPlan(input.PlanID); err != nil {
		return nil, model.NewValidationError("plan_id", "测试计划不存在")
	}
	now := time.Now()
	input.ID = idgen.Hex()
	input.CreatedAt = now
	input.UpdatedAt = now
	next, err := NextRunAfter(input.CronExpr, now)
	if err != nil {
		return nil, model.NewValidationError("cron_expr", "cron 表达式无法解析出下次运行时间")
	}
	input.NextRunAt = &next
	if err := s.store.CreateSchedule(&input); err != nil {
		return nil, err
	}
	s.log.Infof("创建调度 %s(计划=%s, cron=%s, 下次=%s)", input.ID, input.PlanID, input.CronExpr, next.Format(time.RFC3339))
	return &input, nil
}

// TriggerSchedule 手动触发调度：pending -> triggered，并创建一次执行记录。
func (s *Service) TriggerSchedule(id string) (*model.Execution, error) {
	sc, err := s.store.GetSchedule(id)
	if err != nil {
		return nil, err
	}
	if sc.Status == model.ScheduleCancelled {
		return nil, model.NewValidationError("status", "已取消的调度无法触发")
	}
	plan, err := s.store.GetTestPlan(sc.PlanID)
	if err != nil {
		return nil, err
	}
	// 选择在线执行器。
	executor := s.pickOnlineExecutor()
	if executor == nil {
		return nil, model.NewValidationError("executor_id", "没有可用的在线执行器")
	}
	execution, err := s.CreateExecution(model.Execution{PlanID: plan.ID, ExecutorID: executor.ID})
	if err != nil {
		return nil, err
	}
	now := time.Now()
	sc.Status = model.ScheduleTriggered
	sc.LastRunAt = &now
	next, _ := NextRunAfter(sc.CronExpr, now)
	if !next.IsZero() {
		sc.NextRunAt = &next
	}
	sc.UpdatedAt = now
	if err := s.store.UpdateSchedule(sc); err != nil {
		return nil, err
	}
	s.log.Infof("调度 %s 已触发，创建执行记录 %s", sc.ID, execution.ID)
	return execution, nil
}

// pickOnlineExecutor 选择一个在线执行器。
func (s *Service) pickOnlineExecutor() *model.Executor {
	executors := s.store.ListExecutors()
	for _, e := range executors {
		if e.Status == model.ExecutorOnline {
			return e
		}
	}
	return nil
}

// GetSchedule 按 ID 查询调度。
func (s *Service) GetSchedule(id string) (*model.Schedule, error) {
	return s.store.GetSchedule(id)
}

// ListSchedules 分页查询调度。
func (s *Service) ListSchedules(filter model.ScheduleFilter, page, size int) ([]*model.Schedule, int, error) {
	all := s.store.ListSchedules()
	matched := make([]*model.Schedule, 0, len(all))
	for _, sc := range all {
		if filter.Match(sc) {
			matched = append(matched, sc)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []*model.Schedule{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}

// UpdateSchedule 更新调度。
func (s *Service) UpdateSchedule(id string, input model.Schedule) (*model.Schedule, error) {
	exist, err := s.store.GetSchedule(id)
	if err != nil {
		return nil, err
	}
	if input.PlanID != "" {
		exist.PlanID = input.PlanID
	}
	if input.CronExpr != "" {
		exist.CronExpr = input.CronExpr
	}
	if input.Status != "" {
		exist.Status = input.Status
	}
	if input.NextRunAt != nil {
		exist.NextRunAt = input.NextRunAt
	}
	if err := exist.Validate(); err != nil {
		return nil, err
	}
	if _, err := s.store.GetTestPlan(exist.PlanID); err != nil {
		return nil, model.NewValidationError("plan_id", "测试计划不存在")
	}
	exist.UpdatedAt = time.Now()
	if err := s.store.UpdateSchedule(exist); err != nil {
		return nil, err
	}
	return exist, nil
}

// TransitionSchedule 执行调度状态机流转。
func (s *Service) TransitionSchedule(id, to string) (*model.Schedule, error) {
	exist, err := s.store.GetSchedule(id)
	if err != nil {
		return nil, err
	}
	if !model.IsScheduleStatus(to) {
		return nil, model.NewValidationError("status", "调度状态不合法")
	}
	if !model.ScheduleCanTransition(exist.Status, to) {
		return nil, model.NewValidationError("status", "调度状态不允许从 "+exist.Status+" 流转到 "+to)
	}
	exist.Status = to
	now := time.Now()
	exist.UpdatedAt = now
	if to == model.ScheduleTriggered {
		exist.LastRunAt = &now
	}
	if err := s.store.UpdateSchedule(exist); err != nil {
		return nil, err
	}
	return exist, nil
}

// DeleteSchedule 删除调度。
func (s *Service) DeleteSchedule(id string) error {
	if _, err := s.store.GetSchedule(id); err != nil {
		return err
	}
	return s.store.DeleteSchedule(id)
}

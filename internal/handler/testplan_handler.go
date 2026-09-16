package handler

import (
	"net/http"

	"loadtest/internal/model"
	"loadtest/pkg/httpx"
)

// registerTestPlanRoutes 注册测试计划路由。
func (s *Server) registerTestPlanRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/plans", s.createTestPlan)
	mux.HandleFunc("GET /api/plans", s.listTestPlans)
	mux.HandleFunc("GET /api/plans/{id}", s.getTestPlan)
	mux.HandleFunc("PUT /api/plans/{id}", s.updateTestPlan)
	mux.HandleFunc("DELETE /api/plans/{id}", s.deleteTestPlan)
	mux.HandleFunc("POST /api/plans/{id}/transition", s.transitionPlan)
	mux.HandleFunc("POST /api/plans/batch-transition", s.batchTransitionPlans)
}

type createTestPlanRequest struct {
	Name        string `json:"name"`
	ScenarioID  string `json:"scenario_id"`
	Description string `json:"description"`
}

func (s *Server) createTestPlan(w http.ResponseWriter, r *http.Request) {
	var req createTestPlanRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.CreateTestPlan(model.TestPlan{Name: req.Name, ScenarioID: req.ScenarioID, Description: req.Description})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, p)
}

func (s *Server) listTestPlans(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.PlanFilter{
		ScenarioID: r.URL.Query().Get("scenario_id"),
		Status:     r.URL.Query().Get("status"),
		Keyword:    r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListTestPlans(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getTestPlan(w http.ResponseWriter, r *http.Request) {
	p, err := s.svc.GetTestPlan(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

type updateTestPlanRequest struct {
	Name        string `json:"name"`
	ScenarioID  string `json:"scenario_id"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (s *Server) updateTestPlan(w http.ResponseWriter, r *http.Request) {
	var req updateTestPlanRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.UpdateTestPlan(r.PathValue("id"), model.TestPlan{
		Name: req.Name, ScenarioID: req.ScenarioID, Description: req.Description, Status: req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

func (s *Server) deleteTestPlan(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteTestPlan(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

type transitionPlanRequest struct {
	Status string `json:"status"`
}

func (s *Server) transitionPlan(w http.ResponseWriter, r *http.Request) {
	var req transitionPlanRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.TransitionPlan(r.PathValue("id"), req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

type batchTransitionPlansRequest struct {
	IDs    []string `json:"ids"`
	Status string   `json:"status"`
}

func (s *Server) batchTransitionPlans(w http.ResponseWriter, r *http.Request) {
	var req batchTransitionPlansRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	if len(req.IDs) == 0 {
		httpx.BadRequest(w, "计划 ID 列表不能为空")
		return
	}
	items, err := s.svc.BatchTransitionPlans(req.IDs, req.Status)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, items)
}

package handler

import (
	"net/http"

	"loadtest/internal/model"
	"loadtest/pkg/httpx"
)

// registerAssertionRoutes 注册断言路由。
func (s *Server) registerAssertionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/assertions", s.createAssertion)
	mux.HandleFunc("GET /api/assertions", s.listAssertions)
	mux.HandleFunc("GET /api/assertions/{id}", s.getAssertion)
	mux.HandleFunc("DELETE /api/assertions/{id}", s.deleteAssertion)
}

type createAssertionRequest struct {
	ExecutionID string  `json:"execution_id"`
	Metric      string  `json:"metric"`
	Operator    string  `json:"operator"`
	Threshold   float64 `json:"threshold"`
}

func (s *Server) createAssertion(w http.ResponseWriter, r *http.Request) {
	var req createAssertionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	a, err := s.svc.CreateAssertion(model.Assertion{
		ExecutionID: req.ExecutionID, Metric: req.Metric, Operator: req.Operator, Threshold: req.Threshold,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, a)
}

func (s *Server) listAssertions(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.AssertionFilter{
		ExecutionID: r.URL.Query().Get("execution_id"),
		Metric:      r.URL.Query().Get("metric"),
		Result:      r.URL.Query().Get("result"),
	}
	items, total, err := s.svc.ListAssertions(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getAssertion(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetAssertion(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, a)
}

func (s *Server) deleteAssertion(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteAssertion(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

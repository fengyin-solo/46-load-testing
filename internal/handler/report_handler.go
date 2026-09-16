package handler

import (
	"net/http"

	"loadtest/internal/model"
	"loadtest/pkg/httpx"
)

// registerReportRoutes 注册报告路由。
func (s *Server) registerReportRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/reports/generate", s.generateReport)
	mux.HandleFunc("GET /api/reports", s.listReports)
	mux.HandleFunc("GET /api/reports/{id}", s.getReport)
	mux.HandleFunc("GET /api/reports/by-execution/{execution_id}", s.getReportByExecution)
	mux.HandleFunc("DELETE /api/reports/{id}", s.deleteReport)
}

type generateReportRequest struct {
	ExecutionID string `json:"execution_id"`
}

func (s *Server) generateReport(w http.ResponseWriter, r *http.Request) {
	var req generateReportRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	report, err := s.svc.GenerateReport(req.ExecutionID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, report)
}

func (s *Server) listReports(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ReportFilter{
		ExecutionID: r.URL.Query().Get("execution_id"),
		Keyword:     r.URL.Query().Get("keyword"),
	}
	items, total, err := s.svc.ListReports(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{
		Items:      items,
		Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total},
	})
}

func (s *Server) getReport(w http.ResponseWriter, r *http.Request) {
	report, err := s.svc.GetReport(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, report)
}

func (s *Server) getReportByExecution(w http.ResponseWriter, r *http.Request) {
	report, err := s.svc.GetReportByExecution(r.PathValue("execution_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, report)
}

func (s *Server) deleteReport(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteReport(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}

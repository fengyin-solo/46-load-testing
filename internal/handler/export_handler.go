package handler

import (
	"net/http"

	"loadtest/pkg/httpx"
)

// registerExportRoutes 注册数据导出路由。
func (s *Server) registerExportRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/export", s.exportSnapshot)
}

// exportSnapshot 导出汇总快照 JSON。
func (s *Server) exportSnapshot(w http.ResponseWriter, r *http.Request) {
	topN := httpx.ParseQueryInt(r, "top_n", 10)
	snapshot, err := s.svc.ExportSnapshot(topN)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, snapshot)
}

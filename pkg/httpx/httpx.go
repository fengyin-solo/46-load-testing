// Package httpx 提供 HTTP 响应与请求解析的通用工具。
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

// Response 统一 API 响应结构。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON 写入统一格式的 JSON 响应。
func JSON(w http.ResponseWriter, status, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Response{Code: code, Message: message, Data: data})
}

// OK 写入 200 成功响应。
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, 0, "ok", data)
}

// Created 写入 201 创建成功响应。
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, 0, "ok", data)
}

// NoContent 写入 204 无内容响应。
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error 写入错误响应。
func Error(w http.ResponseWriter, status, code int, message string) {
	JSON(w, status, code, message, nil)
}

// BadRequest 写入 400 响应。
func BadRequest(w http.ResponseWriter, message string) { Error(w, http.StatusBadRequest, 400, message) }

// Unauthorized 写入 401 响应。
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, 401, message)
}

// Forbidden 写入 403 响应。
func Forbidden(w http.ResponseWriter, message string) { Error(w, http.StatusForbidden, 403, message) }

// NotFound 写入 404 响应。
func NotFound(w http.ResponseWriter, message string) { Error(w, http.StatusNotFound, 404, message) }

// Conflict 写入 409 响应。
func Conflict(w http.ResponseWriter, message string) { Error(w, http.StatusConflict, 409, message) }

// TooManyRequests 写入 429 响应。
func TooManyRequests(w http.ResponseWriter, message string) {
	Error(w, http.StatusTooManyRequests, 429, message)
}

// InternalError 写入 500 响应。
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, 500, message)
}

// Decode 解析 JSON 请求体，限制 1MB，且只允许单个 JSON 对象。
func Decode(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("请求体只能包含单个 JSON 对象")
	}
	return nil
}

// Pagination 分页信息。
type Pagination struct {
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// PageResult 分页查询结果。
type PageResult struct {
	Items      interface{} `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

// PageParams 分页入参。
type PageParams struct {
	Page int
	Size int
}

// ParsePagination 从查询参数解析分页信息。
func ParsePagination(r *http.Request, defaultSize, maxSize int) PageParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = defaultSize
	}
	if size > maxSize {
		size = maxSize
	}
	return PageParams{Page: page, Size: size}
}

// ParseQueryInt 从查询参数解析整数，失败返回默认值。
func ParseQueryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// ParseQueryFloat 从查询参数解析浮点数，失败返回默认值。
func ParseQueryFloat(r *http.Request, key string, def float64) float64 {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return n
}

# 压测平台（loadtest）

一个纯 Go 标准库（零第三方依赖）实现的压测平台后端服务。围绕「目标服务 → 场景 → 测试计划 → 执行器 → 执行记录 → 指标样本 → 断言 → 报告 → 调度」的完整闭环，提供多状态机、指标聚合统计、断言判定、报告生成、鉴权/限流/请求日志中间件、数据导出与前端看板。

## 技术栈

- Go 1.22（仅 `net/http` + 标准库）
- 内存存储（`sync.RWMutex` 并发安全）
- 原生 HTML + JS + CSS 前端（零 CDN）

## 目录结构

```
origin/
├── cmd/server/main.go          # 入口
├── internal/config/            # 配置
├── internal/app/               # 依赖装配
├── internal/model/             # 实体 + 状态机 + 校验 + 聚合结构
├── internal/store/             # 接口 + 内存实现
├── internal/service/           # 业务逻辑 + 统计
├── internal/handler/           # HTTP 处理 + 中间件 + 静态
├── pkg/httpx  pkg/idgen  pkg/logger
└── web/                        # 前端页面
```

## 运行

```bash
cd origin
go build ./...
go run ./cmd/server
# 默认监听 :8080
```

环境变量：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| PORT | 8080 | 监听端口 |
| ADDR | 空 | 覆盖监听地址 |
| API_KEY | loadtest-secret-key | API 鉴权密钥 |
| RATE_LIMIT | 1000 | 每窗口限流次数 |
| RATE_WINDOW_SEC | 60 | 限流窗口秒数 |
| MAX_PAGE_SIZE | 100 | 分页上限 |
| LOG_LEVEL | info | debug/info/warn/error |

浏览器访问 `http://localhost:8080/` 查看看板。

## 鉴权

除静态页面外，所有 `/api/*` 接口均需请求头 `X-API-Key: <API_KEY>`。

## API 一览

统一响应：`{"code":0,"message":"ok","data":...}`。

### 目标服务 Target

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/targets | 创建目标服务 |
| GET | /api/targets | 列表（method/status/keyword + 分页） |
| GET | /api/targets/{id} | 查询目标服务 |
| PUT | /api/targets/{id} | 更新目标服务 |
| DELETE | /api/targets/{id} | 删除目标服务 |
| PATCH | /api/targets/{id}/status | 更新健康状态 |

### 场景 Scenario

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/scenarios | 创建场景 |
| GET | /api/scenarios | 列表（target_id/method/status/keyword + 分页） |
| GET | /api/scenarios/{id} | 查询场景 |
| PUT | /api/scenarios/{id} | 更新场景 |
| DELETE | /api/scenarios/{id} | 删除场景 |
| POST | /api/scenarios/{id}/transition | 状态流转 |

### 测试计划 TestPlan

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/plans | 创建计划 |
| GET | /api/plans | 列表（scenario_id/status/keyword + 分页） |
| GET | /api/plans/{id} | 查询计划 |
| PUT | /api/plans/{id} | 更新计划 |
| DELETE | /api/plans/{id} | 删除计划 |
| POST | /api/plans/{id}/transition | 状态流转 |
| POST | /api/plans/batch-transition | 批量流转状态 |

### 执行器 Executor

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/executors | 创建执行器 |
| GET | /api/executors | 列表（status/keyword + 分页） |
| GET | /api/executors/{id} | 查询执行器 |
| PUT | /api/executors/{id} | 更新执行器 |
| DELETE | /api/executors/{id} | 删除执行器 |
| POST | /api/executors/{id}/transition | 状态流转 |
| POST | /api/executors/{id}/heartbeat | 心跳 |

### 执行记录 Execution

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/executions | 创建执行记录 |
| GET | /api/executions | 列表（plan_id/executor_id/status + 分页） |
| GET | /api/executions/{id} | 查询执行记录 |
| PUT | /api/executions/{id} | 更新统计 |
| DELETE | /api/executions/{id} | 删除执行记录 |
| POST | /api/executions/{id}/start | 启动（pending→running） |
| POST | /api/executions/{id}/complete | 完成（→completed，生成报告） |
| POST | /api/executions/{id}/fail | 失败（→failed） |
| POST | /api/executions/{id}/cancel | 取消（→cancelled） |

### 指标样本 MetricSample

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/metrics | 创建单条样本 |
| POST | /api/metrics/batch | 批量创建样本 |
| GET | /api/metrics | 列表（execution_id/min_tps/max_error_rate + 分页） |
| GET | /api/metrics/{id} | 查询样本 |
| DELETE | /api/metrics/{id} | 删除样本 |

### 断言 Assertion

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/assertions | 创建断言（自动判定 pass/fail） |
| GET | /api/assertions | 列表（execution_id/metric/result + 分页） |
| GET | /api/assertions/{id} | 查询断言 |
| DELETE | /api/assertions/{id} | 删除断言 |

### 报告 Report

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/reports/generate | 生成报告 |
| GET | /api/reports | 列表（execution_id/keyword + 分页） |
| GET | /api/reports/{id} | 查询报告 |
| GET | /api/reports/by-execution/{execution_id} | 按执行查报告 |
| DELETE | /api/reports/{id} | 删除报告 |

### 调度 Schedule

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/schedules | 创建调度 |
| GET | /api/schedules | 列表（plan_id/status + 分页） |
| GET | /api/schedules/{id} | 查询调度 |
| PUT | /api/schedules/{id} | 更新调度 |
| DELETE | /api/schedules/{id} | 删除调度 |
| POST | /api/schedules/{id}/transition | 状态流转 |

### 统计与导出

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/stats/overview | 全局统计快照 |
| GET | /api/stats/executions/{id} | 单执行聚合 |
| GET | /api/stats/by-plan | 按计划聚合 |
| GET | /api/stats/by-executor | 按执行器聚合 |
| GET | /api/stats/top-scenarios?n=10 | TOP N 场景 |
| GET | /api/export?top_n=10 | 导出汇总快照 |

## 状态机

| 实体 | 流转规则 |
|------|---------|
| Execution | pending→running→completed/failed/cancelled |
| TestPlan | draft→active→paused→archived |
| Executor | online→busy→offline→online |
| Scenario | draft→ready→archived |
| Schedule | pending→triggered→cancelled |

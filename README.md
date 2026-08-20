# cleanroom-monitor

GMP 洁净车间环境监测系统（Go + SQLite + 内嵌前端）。

## 功能

- 环境读数采集（HTTP 上报 + 模拟器 + 采集器 worker）
- 告警规则引擎（阈值 + 持续时长触发，warn/alarm 升级）
- 房间洁净状态机（at_rest / normal / alert / alarm / restricted / release）
- 传感器校准管理与超期标记
- 生产批次关联与批次放行判定
- 趋势报表 / 汇总报表 / CSV 导出
- 实时监控看板（web/ 前端）

## 快速开始

```bash
go build ./...
go run . -- 默认监听 :8080
```

访问 http://localhost:8080/web/index.html

## 接口

见 `internal/handler/router.go`（REST，前缀 `/api/v1`）。

## 测试

```bash
go test ./...
```
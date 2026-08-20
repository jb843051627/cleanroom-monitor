# cleanroom-monitor：GMP 洁净车间环境监测系统

> 垂直领域：制药 / 半导体 GMP（Good Manufacturing Practice）洁净车间环境监测。
> 本系统用于对洁净区（Cleanroom）内各房间（Room）的环境参数（温湿度、压差、粒子计数）进行持续采集、告警评估、洁净状态管理与批次关联，并提供监控看板与报表导出。
> 业务纯为"环境监测 + 状态评估 + 告警"，不涉及库存 / 订单 / 账本 / RBAC / 预约等禁题细分项。

## 1. 核心业务实体

| 实体 | 字段要点 | 说明 |
|---|---|---|
| Cleanroom | id, name, code, grade(ISO5/ISO7/ISO8), area_sqm, status(clean_status), created_at | 洁净区，可含多个房间 |
| Room | id, cleanroom_id, name, code, kind(classify/atelier/locker), area_sqm, target_pressure, pressure_tolerance, target_temp, target_humidity, status | 房间，绑定洁净区 |
| MonitoringPoint | id, room_id, code, param_type(temp/humidity/pressure/particle_05/particle_50), threshold_min, threshold_max, enabled, alarm_duration_sec | 监测点，定义某参数的采集与阈值 |
| Sensor | id, point_id, serial, vendor, last_seen_at, battery, status(active/fault/offline), calibration_due_at | 物理传感器，绑定监测点 |
| Reading | id, point_id, sensor_id, value, param_type, measured_at, raw | 单条环境读数（落盘 SQLite） |
| AlertRule | id, room_id, code, param_type, op(gt/lt/gte/lte), threshold, duration_sec, level(warn/alarm), enabled | 告警规则（阈值持续时长触发） |
| Alert | id, rule_id, room_id, point_id, level, message, status(open/acknowledged/resolved), opened_at, resolved_at | 告警记录 |
| Calibration | id, sensor_id, performed_at, due_at, standard, result, operator | 传感器校准记录 |
| CleanBatch | id, room_id, name, product, phase, start_at, end_at, status(planning/in_progress/completed/aborted) | 生产批次，与房间环境状态关联 |
| CleanStatus | id, room_id, batch_id, state(at_rest/normal/alert/alarm/restricted/release), reason, changed_at | 房间洁净状态机流转记录 |

## 2. 功能模块

1. **采集网关**：传感器/模拟器通过 HTTP `POST /api/v1/readings` 批量上报读数；采集器 Collector 后台轮询网关队列；并发写入 DB。
2. **读数服务**：写入后触发告警评估；提供按房间/时间窗/参数查询、聚合统计（均值/峰值/超标时长）。
3. **告警引擎**：对每条新读数匹配 AlertRule（阈值 + 持续时长），状态机 open → acknowledged → resolved；升级联动（warn 持续升级 alarm）。
4. **洁净状态机**：房间状态 at_rest → normal → alert → alarm → restricted → release；批量完成/中止驱动状态回退。
5. **校准管理**：传感器校准计划、超期提醒、校准记录；过期传感器标记 fault。
6. **批次关联**：生产批次与房间环境状态关联，批次放行（release）要求房间连续满足阈值。
7. **报表与导出**：趋势报表（温度曲线、粒子计数趋势）、汇总报表（房间达标率）、CSV 导出、日报生成。
8. **监控看板（前端）**：web/ 目录下内嵌 HTML/JS/CSS 页面，展示房间实时读数、告警列表、状态徽标（前端不计入 Go 代码统计）。

## 3. HTTP 接口清单（REST，路由前缀 /api/v1）

| 方法 | 路径 | 服务方法 |
|---|---|---|
| POST | /api/v1/readings | ReadingService.Ingest |
| GET | /api/v1/readings/query | ReadingService.Query |
| GET | /api/v1/readings/stats | ReadingService.Stats |
| GET | /api/v1/readings/realtime | ReadingService.RealtimeSnapshot |
| POST | /api/v1/cleanrooms | CleanroomService.Create |
| GET | /api/v1/cleanrooms | CleanroomService.List |
| GET | /api/v1/cleanrooms/{id} | CleanroomService.Get |
| PUT | /api/v1/cleanrooms/{id} | CleanroomService.Update |
| GET | /api/v1/rooms | RoomService.List |
| POST | /api/v1/rooms | RoomService.Create |
| GET | /api/v1/rooms/{id}/status | RoomService.Status |
| PUT | /api/v1/rooms/{id}/status | RoomService.TransitionStatus |
| GET | /api/v1/points | PointService.List |
| POST | /api/v1/points | PointService.Create |
| GET | /api/v1/sensors | SensorService.List |
| POST | /api/v1/sensors | SensorService.Register |
| POST | /api/v1/sensors/{id}/calibrate | CalibrationService.Record |
| GET | /api/v1/sensors/{id}/calibrations | CalibrationService.List |
| POST | /api/v1/alert-rules | RuleService.Create |
| GET | /api/v1/alert-rules | RuleService.List |
| PUT | /api/v1/alert-rules/{id}/toggle | RuleService.Toggle |
| GET | /api/v1/alerts | AlertService.List |
| PUT | /api/v1/alerts/{id}/ack | AlertService.Acknowledge |
| PUT | /api/v1/alerts/{id}/resolve | AlertService.Resolve |
| POST | /api/v1/batches | BatchService.Start |
| PUT | /api/v1/batches/{id}/complete | BatchService.Complete |
| PUT | /api/v1/batches/{id}/abort | BatchService.Abort |
| GET | /api/v1/batches | BatchService.List |
| GET | /api/v1/reports/trend | ReportService.Trend |
| GET | /api/v1/reports/summary | ReportService.Summary |
| GET | /api/v1/reports/export | ReportService.Export |
| GET | /api/v1/dashboard | DashboardService.Snapshot |
| GET | /web/... | WebHandler（前端静态） |

约 32 个接口（档位 B 放宽至 30+ 合规）。

## 4. 技术栈

- Go 1.22（go.mod language version 1.22），单体单进程，标准库 net/http。
- 持久化：modernc.org/sqlite（纯 Go 驱动，文件 `cleanroom.db`，禁止 `:memory:`）。
- 第三方直接依赖：仅 `modernc.org/sqlite`（≤10 约束）。
- 内部包结构（≤12 个子包）：config / model / store / service / handler / collector / alerter / report / util = 9 个。
- 并发：采集器 Collector 后台 goroutine + 读数评估 worker 池 + 看板快照缓存更新，至少 1 处 goroutine/channel。
- 前端：web/ 下 index.html + dashboard.js + style.css（不计入 Go 统计）。

## 5. 目录结构（预期 ≥50 个 .go 文件，不含测试）

```
main.go
internal/config/{config,env,defaults}.go
internal/model/{cleanroom,room,point,sensor,reading,alert,rule,calibration,batch,status,types}.go
internal/store/{db,cleanroom_store,room_store,point_store,sensor_store,reading_store,alert_store,rule_store,calibration_store,batch_store,status_store,cache}.go
internal/service/{cleanroom_svc,room_svc,point_svc,reading_svc,alert_svc,rule_svc,sensor_svc,calibration_svc,batch_svc,report_svc,dashboard_svc,export_svc}.go
internal/handler/{router,cleanroom_handler,room_handler,point_handler,reading_handler,alert_handler,rule_handler,sensor_handler,calibration_handler,batch_handler,report_handler,dashboard_handler,web_handler}.go
internal/collector/{collector,simulator,gateway}.go
internal/alerter/{engine,evaluator,notifier}.go
internal/report/{trend,summary,export}.go
internal/util/{time,math,validator}.go
web/{index.html,dashboard.js,style.css}
```

## 6. 业务规则要点（bug 埋点素材）

- 告警触发：读数超出规则阈值且**持续 ≥ duration_sec**（用最近窗口内连续超标的百分比判定）才开告警。
- 状态机：任意 alarm 未决 → room=alarm；无 alarm 且 30 分钟无读数 → room=restricted；批次完成且房间达标 → release。
- 校准超期：校准 due_at 过期 → sensor 标记 fault，其读数不再参与告警评估。
- 读数入库：同一 (point_id, measured_at) 去重；批量上报部分失败整批拒绝（事务）。
- 聚合统计：按 room 聚合时用参数类型归一化（temp 单位 ℃，humidity %，pressure Pa，particle 个/m³）。
- 达标率：房间达标 = 当小时内读数全部在 [threshold_min, threshold_max] 内的占比。
- 时间处理：所有业务时间以本地时区（Asia/Shanghai, UTC+8）存储与展示，内部统一 int64 Unix 秒。
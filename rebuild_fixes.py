# -*- coding: utf-8 -*-
"""把每个临时修复重建为相对 bug_base_N 的单个、足够深的 commit。"""
import os
import subprocess
import sys

REPO = r"D:\Develop\Workspace-trae\go-workspace\cleanroom-monitor"


def run(cmd, check=True):
    result = subprocess.run(cmd, cwd=REPO, capture_output=True, text=True, shell=True)
    if check and result.returncode != 0:
        print("CMD FAIL:", cmd)
        print(result.stdout)
        print(result.stderr)
        sys.exit(1)
    return result


def apply_edits(edits):
    touched = set()
    for path, old, new in edits:
        full_path = os.path.join(REPO, path)
        with open(full_path, encoding="utf-8") as handle:
            content = handle.read()
        count = content.count(old)
        if count != 1:
            print(f"ANCHOR FAIL {path}: found {count} for {old[:100]!r}")
            sys.exit(1)
        with open(full_path, "w", encoding="utf-8", newline="") as handle:
            handle.write(content.replace(old, new, 1))
        touched.add(path)
    return touched


def write_file(path, content):
    full_path = os.path.join(REPO, path)
    with open(full_path, "w", encoding="utf-8", newline="") as handle:
            handle.write(content)


def commit(message):
    run("git add -A")
    run(f'git commit -q -m "{message}"')


def gofmt(paths):
    for path in sorted(paths):
        run(f'gofmt -w "{path}"')


def load_specs():
    import create_bugs_1 as p1
    import create_bugs_2 as p2
    import create_bugs_3a as p3a
    import create_bugs_3b as p3b

    specs = {}
    for module in (p1, p2, p3a, p3b):
        specs.update(module.BUGS)
    return specs


S = {}

# Cache 修复补充：复制快照的逻辑抽成独立方法，避免返回共享嵌套 slice。
S[2] = [
    ("internal/store/cache.go", "clone.Realtime = make([]model.RealtimeReading, len(s.Realtime))\n\t\tcopy(clone.Realtime, s.Realtime)", "clone.Realtime = cloneRealtime(s.Realtime)"),
    ("internal/store/cache.go", "// Remove 移除房间快照。", "func cloneRealtime(in []model.RealtimeReading) []model.RealtimeReading {\n\tout := make([]model.RealtimeReading, len(in))\n\tcopy(out, in)\n\treturn out\n}\n\n// Remove 移除房间快照。"),
]

S[3] = [
    ("internal/store/cache.go", "\t\tout = append(out, s)", "\t\tout = append(out, cloneRoomSnapshot(s))"),
    ("internal/store/cache.go", "// UpdatedAt 缓存最近刷新时间。", "func cloneRealtime(in []model.RealtimeReading) []model.RealtimeReading {\n\tout := make([]model.RealtimeReading, len(in))\n\tcopy(out, in)\n\treturn out\n}\n\nfunc cloneRoomSnapshot(s *RoomSnapshot) *RoomSnapshot {\n\tif s == nil {\n\t\treturn nil\n\t}\n\tout := *s\n\tout.Realtime = cloneRealtime(s.Realtime)\n\treturn &out\n}\n\n// UpdatedAt 缓存最近刷新时间。"),
]

S[4] = [
    ("internal/service/batch_svc.go", "fmt.Errorf(\"批次不存在: %w\", err)", "wrapBatchError(\"批次不存在\", err)"),
    ("internal/service/batch_svc.go", "fmt.Errorf(\"房间不存在: %w\", err)", "wrapBatchError(\"房间不存在\", err)"),
    ("internal/service/batch_svc.go", "// Abort 中止批次：状态机 → at_rest。", "func wrapBatchError(label string, err error) error {\n\treturn fmt.Errorf(\"%s: %w\", label, err)\n}\n\n// Abort 中止批次：状态机 → at_rest。"),
]

S[5] = [
    ("internal/alerter/engine.go", "\tif e.reads == nil {", "\tif err := ctx.Err(); err != nil {\n\t\treturn false, err\n\t}\n\tif e.reads == nil {"),
]

S[6] = [
    ("internal/service/report_svc.go", "\troom, err := s.rooms.GetByID(ctx, roomID)", "\troom, err := s.loadRoom(ctx, roomID)"),
    ("internal/service/report_svc.go", "// TrendPoint 趋势点。", "func (s *ReportService) loadRoom(ctx context.Context, roomID int64) (*model.Room, error) {\n\troom, err := s.rooms.GetByID(ctx, roomID)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn room, nil\n}\n\n// TrendPoint 趋势点。"),
]

S[7] = [
    ("internal/store/db.go", "\tdefer tx.Rollback()", "\tdefer db.rollback(tx)"),
    ("internal/store/db.go", "// WithTx 事务执行。", "func (db *DB) rollback(tx *sql.Tx) {\n\t_ = tx.Rollback()\n}\n\n// WithTx 事务执行。"),
]

S[8] = [
    ("internal/store/point_store.go", "func (s *SQLPointStore) ListByRoom(ctx context.Context, roomID int64) ([]*model.MonitoringPoint, error) {\n\trows, err := s.db.QueryContext(ctx, \"SELECT \"+pointCols+\" FROM monitoring_points WHERE room_id = ? ORDER BY id\", roomID)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\tdefer rows.Close()\n\treturn scanPoints(rows)\n}", "func (s *SQLPointStore) ListByRoom(ctx context.Context, roomID int64) ([]*model.MonitoringPoint, error) {\n\trows, err := s.db.QueryContext(ctx, \"SELECT \"+pointCols+\" FROM monitoring_points WHERE room_id = ? ORDER BY id\", roomID)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\tdefer rows.Close()\n\tpoints, err := scanPoints(rows)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn clonePoints(points), nil\n}"),
    ("internal/store/point_store.go", "func scanPoints(rows *sql.Rows) ([]*model.MonitoringPoint, error) {", "func clonePoints(in []*model.MonitoringPoint) []*model.MonitoringPoint {\n\tout := make([]*model.MonitoringPoint, len(in))\n\tfor i, p := range in {\n\t\tif p != nil {\n\t\t\tcopyOf := *p\n\t\t\tout[i] = &copyOf\n\t\t}\n\t}\n\treturn out\n}\n\nfunc scanPoints(rows *sql.Rows) ([]*model.MonitoringPoint, error) {"),
]

S[9] = [
    ("internal/service/calibration_svc.go", "\tse, err := s.sensors.GetByID(ctx, in.SensorID)", "\tse, err := s.loadSensor(ctx, in.SensorID)"),
    ("internal/service/calibration_svc.go", "// Record 记录校准并刷新传感器校准时间与状态。", "func (s *CalibrationService) loadSensor(ctx context.Context, id int64) (*model.Sensor, error) {\n\tse, err := s.sensors.GetByID(ctx, id)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn se, nil\n}\n\n// Record 记录校准并刷新传感器校准时间与状态。"),
]

S[10] = [
    ("internal/service/export_svc.go", "util.FormatTime(r.MeasuredAt)", "s.formatExportTime(r.MeasuredAt)"),
    ("internal/service/export_svc.go", "// BuildFilename 生成导出文件名。", "func (s *ExportService) formatExportTime(t time.Time) string {\n\treturn util.FormatTime(t)\n}\n\n// BuildFilename 生成导出文件名。"),
]

S[11] = [
    ("internal/service/reading_svc.go", "\t\ts.mu.Lock()\n\t\tif s.lastSeen == nil {\n\t\t\ts.lastSeen = make(map[int64]time.Time)\n\t\t}\n\t\ts.lastSeen[in.PointID] = measuredAt\n\t\ts.mu.Unlock()", "\t\ts.recordLastSeen(in.PointID, measuredAt)"),
    ("internal/service/reading_svc.go", "// Query 查询读数。", "func (s *ReadingService) recordLastSeen(pointID int64, measuredAt time.Time) {\n\ts.mu.Lock()\n\tdefer s.mu.Unlock()\n\tif s.lastSeen == nil {\n\t\ts.lastSeen = make(map[int64]time.Time)\n\t}\n\ts.lastSeen[pointID] = measuredAt\n}\n\n// Query 查询读数。"),
]

S[12] = [
    ("internal/alerter/engine.go", "\t\tpoint, err := e.points.GetByID(ctx, r.PointID)", "\t\tpoint, err := e.lookupPoint(ctx, r.PointID)"),
    ("internal/alerter/engine.go", "// sustained 判断超标是否持续满 duration_sec。", "func (e *Engine) lookupPoint(ctx context.Context, id int64) (*model.MonitoringPoint, error) {\n\tpoint, err := e.points.GetByID(ctx, id)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn point, nil\n}\n\n// sustained 判断超标是否持续满 duration_sec。"),
]

S[13] = [
    ("internal/model/status.go", "\tallowed, ok := ValidTransitions[from]", "\tif from == StateRestricted && to == StateRelease {\n\t\treturn restrictedReleaseAllowed()\n\t}\n\tallowed, ok := ValidTransitions[from]"),
    ("internal/model/status.go", "// States 所有合法状态。", "func restrictedReleaseAllowed() bool {\n\tallowed, ok := ValidTransitions[StateRestricted]\n\treturn ok && allowed[StateRelease]\n}\n\n// States 所有合法状态。"),
]

S[14] = [
    ("internal/service/dashboard_svc.go", "func (s *DashboardService) Refresh(ctx context.Context) error {\n", "func (s *DashboardService) Refresh(ctx context.Context) error {\n\tif err := checkDashboardContext(ctx); err != nil {\n\t\treturn err\n\t}\n"),
    ("internal/service/dashboard_svc.go", "// DashboardService 看板服务：缓存快照刷新与读取。", "func checkDashboardContext(ctx context.Context) error {\n\treturn ctx.Err()\n}\n\n// DashboardService 看板服务：缓存快照刷新与读取。"),
]

S[15] = [
    ("internal/service/report_svc.go", "\tif len(res.Points) > limit {\n\t\tres.Points = res.Points[len(res.Points)-limit:]\n\t}\n\treturn res, nil", "\tres.Points = s.trimTrendPoints(res.Points, limit)\n\treturn res, nil"),
    ("internal/service/report_svc.go", "// RoomCompliance 房间达标率。", "func (s *ReportService) trimTrendPoints(points []TrendPoint, limit int) []TrendPoint {\n\tif len(points) <= limit {\n\t\treturn points\n\t}\n\treturn points[len(points)-limit:]\n}\n\n// RoomCompliance 房间达标率。"),
]

S[16] = [
    ("internal/service/dashboard_svc.go", "import (\n\t\"context\"", "import (\n\t\"context\"\n\t\"fmt\""),
    ("internal/service/dashboard_svc.go", "\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t\ts.cache.Set(r.ID, snap)", "\t\tif err != nil {\n\t\t\treturn fmt.Errorf(\"构建房间快照: %w\", err)\n\t\t}\n\t\ts.cache.Set(r.ID, snap)"),
    ("internal/service/dashboard_svc.go", "\tif err != nil {\n\t\treturn nil, err\n\t}\n\tvar realtime []model.RealtimeReading", "\tif err != nil {\n\t\treturn nil, fmt.Errorf(\"查询房间点位: %w\", err)\n\t}\n\tvar realtime []model.RealtimeReading"),
]

S[17] = [
    ("internal/alerter/evaluator.go", "\tselect {\n\tcase e.queue <- readings:\n\tdefault:\n\t\t// 队列满：丢弃（背压保护，防内存膨胀）\n\t}", "\te.enqueue(readings)"),
    ("internal/alerter/evaluator.go", "// Close 关闭 worker 池。", "func (e *Evaluator) enqueue(readings []*model.Reading) {\n\tselect {\n\tcase e.queue <- readings:\n\tdefault:\n\t\t// 队列满：丢弃（背压保护，防内存膨胀）\n\t}\n}\n\n// Close 关闭 worker 池。"),
]

S[18] = [
    ("internal/service/batch_svc.go", "func (s *BatchService) Complete(ctx context.Context, id int64) (*model.CleanBatch, error) {\n\tb, err := s.batches.GetByID(ctx, id)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\tif b.Status != model.BatchInProgress && b.Status != model.BatchPlanning {\n\t\treturn nil, model.ErrConflict\n\t}\n\troom, err := s.rooms.GetByID(ctx, b.RoomID)", "func (s *BatchService) Complete(ctx context.Context, id int64) (*model.CleanBatch, error) {\n\tb, err := s.batches.GetByID(ctx, id)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\tif b.Status != model.BatchInProgress && b.Status != model.BatchPlanning {\n\t\treturn nil, model.ErrConflict\n\t}\n\troom, err := s.lookupBatchRoom(ctx, b.RoomID)"),
    ("internal/service/batch_svc.go", "// Abort 中止批次：状态机 → at_rest。", "func (s *BatchService) lookupBatchRoom(ctx context.Context, id int64) (*model.Room, error) {\n\troom, err := s.rooms.GetByID(ctx, id)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\treturn room, nil\n}\n\n// Abort 中止批次：状态机 → at_rest。"),
]

S[19] = [
    ("internal/service/batch_svc.go", "\t\t_ = s.batches.UpdateStatus(ctx, b.ID, model.BatchAborted, &[]time.Time{time.Now()}[0])", "\t\ts.abortStartedBatch(ctx, b.ID)"),
    ("internal/service/batch_svc.go", "// Complete 完成批次：状态机 → release（若房间无未决 alarm）否则 restricted。", "func (s *BatchService) abortStartedBatch(ctx context.Context, id int64) {\n\tendAt := time.Now()\n\t_ = s.batches.UpdateStatus(ctx, id, model.BatchAborted, &endAt)\n}\n\n// Complete 完成批次：状态机 → release（若房间无未决 alarm）否则 restricted。"),
]

S[20] = [
    ("internal/service/report_svc.go", "\tfor _, r := range rooms {\n", "\tfor _, r := range rooms {\n\t\tif err := checkReportContext(ctx); err != nil {\n\t\t\treturn nil, err\n\t\t}\n"),
    ("internal/service/report_svc.go", "// StateDistribution 房间状态分布（看板用）。", "func checkReportContext(ctx context.Context) error {\n\treturn ctx.Err()\n}\n\n// StateDistribution 房间状态分布（看板用）。"),
]

S[21] = [
    ("internal/alerter/engine.go", "if !errors.Is(err, model.ErrNotFound) {", "if !isMissingAlert(err) {"),
    ("internal/alerter/engine.go", "// buildMessage 构造告警消息。", "func isMissingAlert(err error) bool {\n\treturn errors.Is(err, model.ErrNotFound)\n}\n\n// buildMessage 构造告警消息。"),
]

S[22] = [
    ("internal/store/alert_store.go", "\treturn out, rows.Err()\n}\n\nfunc (s *SQLAlertStore) ListOpenByRoom", "\treturn cloneAlerts(out), rows.Err()\n}\n\nfunc (s *SQLAlertStore) ListOpenByRoom"),
    ("internal/store/alert_store.go", "func (s *SQLAlertStore) ListOpenByRoom", "func cloneAlerts(in []*model.Alert) []*model.Alert {\n\tout := make([]*model.Alert, len(in))\n\tfor i, alert := range in {\n\t\tif alert != nil {\n\t\t\tcopyOf := *alert\n\t\t\tout[i] = &copyOf\n\t\t}\n\t}\n\treturn out\n}\n\nfunc (s *SQLAlertStore) ListOpenByRoom"),
]

S[25] = [
    ("internal/alerter/engine.go", "\tfrom := now.Add(-time.Duration(rule.DurationSec) * time.Second)", "\tfrom := e.windowStart(now, rule.DurationSec)"),
    ("internal/alerter/engine.go", "// openOrUpdate 开新告警或复用已有告警（避免重复创建）。", "func (e *Engine) windowStart(now time.Time, seconds int) time.Time {\n\treturn now.Add(-time.Duration(seconds) * time.Second)\n}\n\n// openOrUpdate 开新告警或复用已有告警（避免重复创建）。"),
]

S[26] = [
    ("internal/service/export_svc.go", "func (s *ExportService) ExportReadingsCSV(ctx context.Context, roomID int64, from, to time.Time) ([]byte, error) {\n", "func (s *ExportService) ExportReadingsCSV(ctx context.Context, roomID int64, from, to time.Time) ([]byte, error) {\n\tif err := checkExportContext(ctx); err != nil {\n\t\treturn nil, err\n\t}\n"),
    ("internal/service/export_svc.go", "// BuildFilename 生成导出文件名。", "func checkExportContext(ctx context.Context) error {\n\treturn ctx.Err()\n}\n\n// BuildFilename 生成导出文件名。"),
]

S[29] = [
    ("internal/service/batch_svc.go", "\topenAlarm, err := s.alerts.CountOpenByLevel(ctx, b.RoomID, model.AlertAlarm)", "\topenAlarm, err := s.alarmCountForRelease(ctx, b.RoomID)"),
    ("internal/service/batch_svc.go", "// Abort 中止批次：状态机 → at_rest。", "func (s *BatchService) alarmCountForRelease(ctx context.Context, roomID int64) (int, error) {\n\treturn s.alerts.CountOpenByLevel(ctx, roomID, model.AlertAlarm)\n}\n\n// Abort 中止批次：状态机 → at_rest。"),
]


def main():
    specs = load_specs()
    run("git checkout -q main")
    hashes = {}
    for number in range(1, 31):
        rebuild = f"rebuild_fix_{number}"
        run(f"git branch -D {rebuild}", check=False)
        run(f"git checkout -q -b {rebuild} bug_base_{number}")
        touched = set(apply_edits(specs[number]["fix_edits"]))
        touched.update(apply_edits(S.get(number, [])))
        gofmt(touched)
        commit("refactor: internal adjustment")
        hashes[number] = run("git rev-parse HEAD").stdout.strip()
        run("git checkout -q main")
        run(f"git branch -D fix_calib_{number}")
        run(f"git branch -m {rebuild} fix_calib_{number}")
    for number in range(1, 31):
        print(f"bug-{number:03d} fix_calib hash: {hashes[number]}")


if __name__ == "__main__":
    main()

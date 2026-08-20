package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"cleanroom-monitor/internal/model"
	"cleanroom-monitor/internal/service"
)

// Handler 聚合全部 HTTP handler。
type Handler struct {
	services *service.Services
	webDir   string
}

// NewHandler 创建 HTTP handler 聚合。
func NewHandler(services *service.Services, webDir string) *Handler {
	return &Handler{services: services, webDir: webDir}
}

// Router 构建路由。
func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// 读数
	mux.HandleFunc("POST /api/v1/readings", h.gatewayIngest)
	mux.HandleFunc("GET /api/v1/readings/query", h.readingsQuery)
	mux.HandleFunc("GET /api/v1/readings/stats", h.readingsStats)
	mux.HandleFunc("GET /api/v1/readings/realtime", h.readingsRealtime)

	// 洁净区 / 房间 / 点位
	mux.HandleFunc("POST /api/v1/cleanrooms", h.cleanroomCreate)
	mux.HandleFunc("GET /api/v1/cleanrooms", h.cleanroomList)
	mux.HandleFunc("GET /api/v1/cleanrooms/{id}", h.cleanroomGet)
	mux.HandleFunc("PUT /api/v1/cleanrooms/{id}", h.cleanroomUpdate)
	mux.HandleFunc("GET /api/v1/rooms", h.roomList)
	mux.HandleFunc("POST /api/v1/rooms", h.roomCreate)
	mux.HandleFunc("GET /api/v1/rooms/{id}", h.roomGet)
	mux.HandleFunc("GET /api/v1/rooms/{id}/status", h.roomStatus)
	mux.HandleFunc("POST /api/v1/points", h.pointCreate)
	mux.HandleFunc("GET /api/v1/points", h.pointList)
	mux.HandleFunc("PUT /api/v1/points/{id}/toggle", h.pointToggle)

	// 传感器 / 校准
	mux.HandleFunc("POST /api/v1/sensors", h.sensorRegister)
	mux.HandleFunc("GET /api/v1/sensors", h.sensorList)
	mux.HandleFunc("GET /api/v1/sensors/{id}", h.sensorGet)
	mux.HandleFunc("POST /api/v1/sensors/{id}/calibrate", h.calibrationRecord)
	mux.HandleFunc("GET /api/v1/sensors/{id}/calibrations", h.calibrationList)

	// 告警规则 / 告警
	mux.HandleFunc("POST /api/v1/alert-rules", h.ruleCreate)
	mux.HandleFunc("GET /api/v1/alert-rules", h.ruleList)
	mux.HandleFunc("PUT /api/v1/alert-rules/{id}/toggle", h.ruleToggle)
	mux.HandleFunc("GET /api/v1/alerts", h.alertList)
	mux.HandleFunc("PUT /api/v1/alerts/{id}/ack", h.alertAck)
	mux.HandleFunc("PUT /api/v1/alerts/{id}/resolve", h.alertResolve)

	// 批次
	mux.HandleFunc("POST /api/v1/batches", h.batchStart)
	mux.HandleFunc("PUT /api/v1/batches/{id}/complete", h.batchComplete)
	mux.HandleFunc("PUT /api/v1/batches/{id}/abort", h.batchAbort)
	mux.HandleFunc("GET /api/v1/batches", h.batchList)

	// 报表 / 看板
	mux.HandleFunc("GET /api/v1/reports/trend", h.reportTrend)
	mux.HandleFunc("GET /api/v1/reports/summary", h.reportSummary)
	mux.HandleFunc("GET /api/v1/reports/export", h.reportExport)
	mux.HandleFunc("GET /api/v1/dashboard", h.dashboard)

	// 健康检查
	mux.HandleFunc("GET /healthz", h.health)

	// 前端
	fs := http.FileServer(http.Dir(h.webDir))
	mux.HandleFunc("GET /web/", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/web/", fs).ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/index.html", http.StatusFound)
	})

	return logMiddleware(mux)
}

// logMiddleware 请求日志。
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[http] %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// writeJSON 写 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeErr 写错误响应。
func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

// pathID 解析路径参数为 int64。
func pathID(r *http.Request, name string) (int64, error) {
	v := r.PathValue(name)
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil || id <= 0 {
		return 0, model.ErrInvalidInput
	}
	return id, nil
}

// statusFromError 错误码映射。
func statusFromError(err error) int {
	switch err {
	case model.ErrNotFound:
		return http.StatusNotFound
	case model.ErrDuplicateCode, model.ErrConflict, model.ErrStateConflict, model.ErrBatchActive:
		return http.StatusConflict
	case model.ErrInvalidInput, model.ErrNameRequired, model.ErrCodeRequired,
		model.ErrInvalidGrade, model.ErrInvalidArea, model.ErrRoomRequired,
		model.ErrCleanroomRequired, model.ErrInvalidKind, model.ErrInvalidPressureTarget,
		model.ErrPointRequired, model.ErrInvalidParamType, model.ErrThresholdInverted,
		model.ErrInvalidDuration, model.ErrSerialRequired, model.ErrVendorRequired,
		model.ErrInvalidBattery, model.ErrInvalidLevel, model.ErrInvalidAlertStatus,
		model.ErrInvalidOp, model.ErrSensorRequired, model.ErrStandardRequired,
		model.ErrInvalidResult, model.ErrOperatorRequired, model.ErrInvalidDueDate,
		model.ErrProductRequired, model.ErrInvalidBatchStatus, model.ErrInvalidState:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
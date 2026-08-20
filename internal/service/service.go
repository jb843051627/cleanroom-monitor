package service

import (
	"cleanroom-monitor/internal/alerter"
	"cleanroom-monitor/internal/config"
	"cleanroom-monitor/internal/store"
)

// Services 聚合所有业务服务与依赖。
type Services struct {
	Cleanrooms    *CleanroomService
	Rooms         *RoomService
	Points        *PointService
	Readings      *ReadingService
	Alerts        *AlertService
	Rules         *RuleService
	Sensors       *SensorService
	Calibrations  *CalibrationService
	Batches       *BatchService
	Reports       *ReportService
	Dashboard     *DashboardService
	Export        *ExportService

	Store   *store.DB
	Cache   *store.Cache
	Config  *config.AppConfig
	Engine  *alerter.Engine
}

// NewServices 组装全部服务。
func NewServices(db *store.DB, cfg *config.AppConfig) *Services {
	cleanroomStore := store.NewCleanroomStore(db)
	roomStore := store.NewRoomStore(db)
	pointStore := store.NewPointStore(db)
	sensorStore := store.NewSensorStore(db)
	readingStore := store.NewReadingStore(db)
	alertStore := store.NewAlertStore(db)
	ruleStore := store.NewRuleStore(db)
	calStore := store.NewCalibrationStore(db)
	batchStore := store.NewBatchStore(db)
	statusStore := store.NewStatusStore(db)

	cache := store.NewCache()
	engine := alerter.NewEngine(ruleStore, alertStore, pointStore)

	return &Services{
		Cleanrooms:   NewCleanroomService(cleanroomStore),
		Rooms:        NewRoomService(roomStore, cleanroomStore),
		Points:       NewPointService(pointStore, roomStore),
		Readings:     NewReadingService(readingStore, pointStore, sensorStore, alertStore, cache, engine),
		Alerts:       NewAlertService(alertStore),
		Rules:        NewRuleService(ruleStore, roomStore),
		Sensors:      NewSensorService(sensorStore, pointStore),
		Calibrations: NewCalibrationService(calStore, sensorStore),
		Batches:      NewBatchService(batchStore, roomStore, statusStore, alertStore),
		Reports:      NewReportService(readingStore, pointStore, roomStore, cleanroomStore, alertStore, statusStore),
		Dashboard:    NewDashboardService(roomStore, pointStore, cache, alertStore, readingStore),
		Export:       NewExportService(readingStore, pointStore, roomStore),

		Store:  db,
		Cache:  cache,
		Config: cfg,
		Engine: engine,
	}
}
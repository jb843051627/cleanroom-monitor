package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"cleanroom-monitor/internal/model"
)

// PointStore 监测点存储。
type PointStore interface {
	Create(ctx context.Context, p *model.MonitoringPoint) error
	GetByID(ctx context.Context, id int64) (*model.MonitoringPoint, error)
	GetByCode(ctx context.Context, code string) (*model.MonitoringPoint, error)
	ListByRoom(ctx context.Context, roomID int64) ([]*model.MonitoringPoint, error)
	ListByParam(ctx context.Context, paramType string) ([]*model.MonitoringPoint, error)
	ListAll(ctx context.Context) ([]*model.MonitoringPoint, error)
	Update(ctx context.Context, p *model.MonitoringPoint) error
	SetEnabled(ctx context.Context, id int64, enabled bool) error
}

type SQLPointStore struct {
	db *DB
}

func NewPointStore(db *DB) PointStore {
	return &SQLPointStore{db: db}
}

const pointCols = "id, room_id, code, param_type, threshold_min, threshold_max, alarm_duration_sec, enabled, created_at"

func scanPoint(row interface{ Scan(...any) error }) (*model.MonitoringPoint, error) {
	var p model.MonitoringPoint
	var createdAt time.Time
	var enabled int
	if err := row.Scan(&p.ID, &p.RoomID, &p.Code, &p.ParamType, &p.ThresholdMin,
		&p.ThresholdMax, &p.AlarmDurationSec, &enabled, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	p.Enabled = enabled == 1
	p.CreatedAt = createdAt
	return &p, nil
}

func (s *SQLPointStore) Create(ctx context.Context, p *model.MonitoringPoint) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO monitoring_points (room_id, code, param_type, threshold_min, threshold_max, alarm_duration_sec, enabled, created_at) VALUES (?,?,?,?,?,?,?,?)`,
		p.RoomID, p.Code, p.ParamType, p.ThresholdMin, p.ThresholdMax, p.AlarmDurationSec, boolToInt(p.Enabled), p.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrDuplicateCode
		}
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = id
	return nil
}

func (s *SQLPointStore) GetByID(ctx context.Context, id int64) (*model.MonitoringPoint, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+pointCols+" FROM monitoring_points WHERE id = ?", id)
	return scanPoint(row)
}

func (s *SQLPointStore) GetByCode(ctx context.Context, code string) (*model.MonitoringPoint, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+pointCols+" FROM monitoring_points WHERE code = ?", code)
	return scanPoint(row)
}

func (s *SQLPointStore) ListByRoom(ctx context.Context, roomID int64) ([]*model.MonitoringPoint, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+pointCols+" FROM monitoring_points WHERE room_id = ? ORDER BY id", roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPoints(rows)
}

func (s *SQLPointStore) ListByParam(ctx context.Context, paramType string) ([]*model.MonitoringPoint, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+pointCols+" FROM monitoring_points WHERE param_type = ? AND enabled = 1 ORDER BY id", paramType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPoints(rows)
}

func (s *SQLPointStore) ListAll(ctx context.Context) ([]*model.MonitoringPoint, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+pointCols+" FROM monitoring_points ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPoints(rows)
}

func (s *SQLPointStore) Update(ctx context.Context, p *model.MonitoringPoint) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE monitoring_points SET code=?, param_type=?, threshold_min=?, threshold_max=?, alarm_duration_sec=?, enabled=? WHERE id=?`,
		p.Code, p.ParamType, p.ThresholdMin, p.ThresholdMax, p.AlarmDurationSec, boolToInt(p.Enabled), p.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (s *SQLPointStore) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	res, err := s.db.ExecContext(ctx, "UPDATE monitoring_points SET enabled=? WHERE id=?", boolToInt(enabled), id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func scanPoints(rows *sql.Rows) ([]*model.MonitoringPoint, error) {
	var out []*model.MonitoringPoint
	for rows.Next() {
		var p model.MonitoringPoint
		var createdAt time.Time
		var enabled int
		if err := rows.Scan(&p.ID, &p.RoomID, &p.Code, &p.ParamType, &p.ThresholdMin,
			&p.ThresholdMax, &p.AlarmDurationSec, &enabled, &createdAt); err != nil {
			return nil, err
		}
		p.Enabled = enabled == 1
		p.CreatedAt = createdAt
		out = append(out, &p)
	}
	return out, rows.Err()
}
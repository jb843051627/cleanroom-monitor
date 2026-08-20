package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"cleanroom-monitor/internal/model"
)

// AlertStore 告警存储。
type AlertStore interface {
	Create(ctx context.Context, a *model.Alert) error
	GetByID(ctx context.Context, id int64) (*model.Alert, error)
	List(ctx context.Context, in model.AlertInput) ([]*model.Alert, error)
	ListOpenByRoom(ctx context.Context, roomID int64) ([]*model.Alert, error)
	OpenByRulePoint(ctx context.Context, ruleID, pointID int64) (*model.Alert, error)
	SetAck(ctx context.Context, id int64, at time.Time) error
	SetResolved(ctx context.Context, id int64, at time.Time) error
	CountOpenByRoom(ctx context.Context, roomID int64) (int, error)
	CountOpenByLevel(ctx context.Context, roomID int64, level string) (int, error)
}

type SQLAlertStore struct {
	db *DB
}

func NewAlertStore(db *DB) AlertStore {
	return &SQLAlertStore{db: db}
}

const alertCols = "id, rule_id, room_id, point_id, level, message, status, opened_at, ack_at, resolved_at"

func scanAlert(row interface{ Scan(...any) error }) (*model.Alert, error) {
	var a model.Alert
	var ack, resolved sql.NullTime
	var msg sql.NullString
	if err := row.Scan(&a.ID, &a.RuleID, &a.RoomID, &a.PointID, &a.Level, &msg, &a.Status,
		&a.OpenedAt, &ack, &resolved); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	a.Message = msg.String
	if ack.Valid {
		t := ack.Time
		a.AckAt = &t
	}
	if resolved.Valid {
		t := resolved.Time
		a.ResolvedAt = &t
	}
	return &a, nil
}

func (s *SQLAlertStore) Create(ctx context.Context, a *model.Alert) error {
	var ack, resolved sql.NullTime
	if a.AckAt != nil {
		ack.Time = *a.AckAt
		ack.Valid = true
	}
	if a.ResolvedAt != nil {
		resolved.Time = *a.ResolvedAt
		resolved.Valid = true
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO alerts (rule_id, room_id, point_id, level, message, status, opened_at, ack_at, resolved_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		a.RuleID, a.RoomID, a.PointID, a.Level, a.Message, a.Status, a.OpenedAt, ack, resolved)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	a.ID = id
	return nil
}

func (s *SQLAlertStore) GetByID(ctx context.Context, id int64) (*model.Alert, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+alertCols+" FROM alerts WHERE id = ?", id)
	return scanAlert(row)
}

func (s *SQLAlertStore) List(ctx context.Context, in model.AlertInput) ([]*model.Alert, error) {
	q := "SELECT " + alertCols + " FROM alerts WHERE 1=1"
	args := []any{}
	if in.RoomID > 0 {
		q += " AND room_id = ?"
		args = append(args, in.RoomID)
	}
	if in.Level != "" {
		q += " AND level = ?"
		args = append(args, in.Level)
	}
	if in.Status != "" {
		q += " AND status = ?"
		args = append(args, in.Status)
	}
	if !in.From.IsZero() {
		q += " AND opened_at >= ?"
		args = append(args, in.From)
	}
	if !in.To.IsZero() {
		q += " AND opened_at <= ?"
		args = append(args, in.To)
	}
	q += " ORDER BY opened_at DESC"
	if in.Limit <= 0 || in.Limit > 500 {
		in.Limit = 100
	}
	q += " LIMIT ?"
	args = append(args, in.Limit)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Alert
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *SQLAlertStore) ListOpenByRoom(ctx context.Context, roomID int64) ([]*model.Alert, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT "+alertCols+" FROM alerts WHERE room_id=? AND status IN ('open','acknowledged') ORDER BY opened_at",
		roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Alert
	for rows.Next() {
		a, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *SQLAlertStore) OpenByRulePoint(ctx context.Context, ruleID, pointID int64) (*model.Alert, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT "+alertCols+" FROM alerts WHERE rule_id=? AND point_id=? AND status IN ('open','acknowledged') ORDER BY opened_at LIMIT 1",
		ruleID, pointID)
	a, err := scanAlert(row)
	if err != nil {
		return nil, fmt.Errorf("查询未决告警: %w", err)
	}
	return a, nil
}

func (s *SQLAlertStore) SetAck(ctx context.Context, id int64, at time.Time) error {
	res, err := s.db.ExecContext(ctx, "UPDATE alerts SET status='acknowledged', ack_at=? WHERE id=?", at, id)
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

func (s *SQLAlertStore) SetResolved(ctx context.Context, id int64, at time.Time) error {
	res, err := s.db.ExecContext(ctx, "UPDATE alerts SET status='resolved', resolved_at=? WHERE id=?", at, id)
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

func (s *SQLAlertStore) CountOpenByRoom(ctx context.Context, roomID int64) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM alerts WHERE room_id=? AND status IN ('open','acknowledged')", roomID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (s *SQLAlertStore) CountOpenByLevel(ctx context.Context, roomID int64, level string) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM alerts WHERE room_id=? AND status IN ('open','acknowledged') AND level=?", roomID, level).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"cleanroom-monitor/internal/model"
)

// RuleStore 告警规则存储。
type RuleStore interface {
	Create(ctx context.Context, r *model.AlertRule) error
	GetByID(ctx context.Context, id int64) (*model.AlertRule, error)
	GetByCode(ctx context.Context, code string) (*model.AlertRule, error)
	ListByRoom(ctx context.Context, roomID int64) ([]*model.AlertRule, error)
	ListEnabled(ctx context.Context) ([]*model.AlertRule, error)
	ListAll(ctx context.Context) ([]*model.AlertRule, error)
	Update(ctx context.Context, r *model.AlertRule) error
	SetEnabled(ctx context.Context, id int64, enabled bool) error
}

type SQLRuleStore struct {
	db *DB
}

func NewRuleStore(db *DB) RuleStore {
	return &SQLRuleStore{db: db}
}

const ruleCols = "id, room_id, code, param_type, op, threshold, duration_sec, level, enabled, created_at"

func scanRule(row interface{ Scan(...any) error }) (*model.AlertRule, error) {
	var r model.AlertRule
	var createdAt time.Time
	var enabled int
	if err := row.Scan(&r.ID, &r.RoomID, &r.Code, &r.ParamType, &r.Op, &r.Threshold,
		&r.DurationSec, &r.Level, &enabled, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	r.Enabled = enabled == 1
	r.CreatedAt = createdAt
	return &r, nil
}

func (s *SQLRuleStore) Create(ctx context.Context, r *model.AlertRule) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO alert_rules (room_id, code, param_type, op, threshold, duration_sec, level, enabled, created_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		r.RoomID, r.Code, r.ParamType, r.Op, r.Threshold, r.DurationSec, r.Level, boolToInt(r.Enabled), r.CreatedAt)
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
	r.ID = id
	return nil
}

func (s *SQLRuleStore) GetByID(ctx context.Context, id int64) (*model.AlertRule, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+ruleCols+" FROM alert_rules WHERE id = ?", id)
	return scanRule(row)
}

func (s *SQLRuleStore) GetByCode(ctx context.Context, code string) (*model.AlertRule, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+ruleCols+" FROM alert_rules WHERE code = ?", code)
	return scanRule(row)
}

func (s *SQLRuleStore) ListByRoom(ctx context.Context, roomID int64) ([]*model.AlertRule, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+ruleCols+" FROM alert_rules WHERE room_id=? ORDER BY id", roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRules(rows)
}

func (s *SQLRuleStore) ListEnabled(ctx context.Context) ([]*model.AlertRule, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+ruleCols+" FROM alert_rules WHERE enabled=1 ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRules(rows)
}

func (s *SQLRuleStore) ListAll(ctx context.Context) ([]*model.AlertRule, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+ruleCols+" FROM alert_rules ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRules(rows)
}

func (s *SQLRuleStore) Update(ctx context.Context, r *model.AlertRule) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE alert_rules SET code=?, param_type=?, op=?, threshold=?, duration_sec=?, level=?, enabled=? WHERE id=?`,
		r.Code, r.ParamType, r.Op, r.Threshold, r.DurationSec, r.Level, boolToInt(r.Enabled), r.ID)
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

func (s *SQLRuleStore) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	res, err := s.db.ExecContext(ctx, "UPDATE alert_rules SET enabled=? WHERE id=?", boolToInt(enabled), id)
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

func scanRules(rows *sql.Rows) ([]*model.AlertRule, error) {
	var out []*model.AlertRule
	for rows.Next() {
		r, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
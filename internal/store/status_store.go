package store

import (
	"context"
	"database/sql"
	"errors"

	"cleanroom-monitor/internal/model"
)

// StatusStore 洁净状态流转记录存储。
type StatusStore interface {
	Append(ctx context.Context, s *model.CleanStatus) error
	ListByRoom(ctx context.Context, roomID int64, limit int) ([]*model.CleanStatus, error)
	LatestByRoom(ctx context.Context, roomID int64) (*model.CleanStatus, error)
	CountByState(ctx context.Context, state string) (int, error)
}

type SQLStatusStore struct {
	db *DB
}

func NewStatusStore(db *DB) StatusStore {
	return &SQLStatusStore{db: db}
}

const statusCols = "id, room_id, batch_id, state, reason, changed_at"

func scanStatus(row interface{ Scan(...any) error }) (*model.CleanStatus, error) {
	var s model.CleanStatus
	var reason sql.NullString
	if err := row.Scan(&s.ID, &s.RoomID, &s.BatchID, &s.State, &reason, &s.ChangedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	s.Reason = reason.String
	return &s, nil
}

func (s *SQLStatusStore) Append(ctx context.Context, st *model.CleanStatus) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO clean_status (room_id, batch_id, state, reason, changed_at) VALUES (?,?,?,?,?)`,
		st.RoomID, st.BatchID, st.State, st.Reason, st.ChangedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	st.ID = id
	return nil
}

func (s *SQLStatusStore) ListByRoom(ctx context.Context, roomID int64, limit int) ([]*model.CleanStatus, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, "SELECT "+statusCols+" FROM clean_status WHERE room_id=? ORDER BY changed_at DESC LIMIT ?", roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.CleanStatus
	for rows.Next() {
		st, err := scanStatus(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (s *SQLStatusStore) LatestByRoom(ctx context.Context, roomID int64) (*model.CleanStatus, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+statusCols+" FROM clean_status WHERE room_id=? ORDER BY changed_at DESC LIMIT 1", roomID)
	return scanStatus(row)
}

func (s *SQLStatusStore) CountByState(ctx context.Context, state string) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM clean_status WHERE state=?", state).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
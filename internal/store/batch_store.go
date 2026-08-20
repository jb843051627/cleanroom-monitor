package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"cleanroom-monitor/internal/model"
)

// BatchStore 批次存储。
type BatchStore interface {
	Create(ctx context.Context, b *model.CleanBatch) error
	GetByID(ctx context.Context, id int64) (*model.CleanBatch, error)
	GetActiveByRoom(ctx context.Context, roomID int64) (*model.CleanBatch, error)
	ListByRoom(ctx context.Context, roomID int64) ([]*model.CleanBatch, error)
	List(ctx context.Context, limit, offset int) ([]*model.CleanBatch, error)
	UpdateStatus(ctx context.Context, id int64, status string, endAt *time.Time) error
	CountActiveByRoom(ctx context.Context, roomID int64) (int, error)
}

type SQLBatchStore struct {
	db *DB
}

func NewBatchStore(db *DB) BatchStore {
	return &SQLBatchStore{db: db}
}

const batchCols = "id, room_id, name, product, phase, start_at, end_at, status, created_at"

func scanBatch(row interface{ Scan(...any) error }) (*model.CleanBatch, error) {
	var b model.CleanBatch
	var endAt, createdAt sql.NullTime
	if err := row.Scan(&b.ID, &b.RoomID, &b.Name, &b.Product, &b.Phase, &b.StartAt,
		&endAt, &b.Status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	if endAt.Valid {
		t := endAt.Time
		b.EndAt = &t
	}
	b.CreatedAt = createdAt.Time
	return &b, nil
}

func (s *SQLBatchStore) Create(ctx context.Context, b *model.CleanBatch) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO clean_batches (room_id, name, product, phase, start_at, end_at, status, created_at) VALUES (?,?,?,?,?,?,?,?)`,
		b.RoomID, b.Name, b.Product, b.Phase, b.StartAt, b.EndAt, b.Status, b.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	b.ID = id
	return nil
}

func (s *SQLBatchStore) GetByID(ctx context.Context, id int64) (*model.CleanBatch, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+batchCols+" FROM clean_batches WHERE id = ?", id)
	return scanBatch(row)
}

func (s *SQLBatchStore) GetActiveByRoom(ctx context.Context, roomID int64) (*model.CleanBatch, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT "+batchCols+" FROM clean_batches WHERE room_id=? AND status IN ('planning','in_progress') ORDER BY id DESC LIMIT 1", roomID)
	return scanBatch(row)
}

func (s *SQLBatchStore) ListByRoom(ctx context.Context, roomID int64) ([]*model.CleanBatch, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+batchCols+" FROM clean_batches WHERE room_id=? ORDER BY start_at DESC", roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.CleanBatch
	for rows.Next() {
		b, err := scanBatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *SQLBatchStore) List(ctx context.Context, limit, offset int) ([]*model.CleanBatch, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+batchCols+" FROM clean_batches ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.CleanBatch
	for rows.Next() {
		b, err := scanBatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *SQLBatchStore) UpdateStatus(ctx context.Context, id int64, status string, endAt *time.Time) error {
	res, err := s.db.ExecContext(ctx, "UPDATE clean_batches SET status=?, end_at=? WHERE id=?", status, endAt, id)
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

func (s *SQLBatchStore) CountActiveByRoom(ctx context.Context, roomID int64) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM clean_batches WHERE room_id=? AND status IN ('planning','in_progress')", roomID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
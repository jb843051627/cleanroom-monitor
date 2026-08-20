package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"cleanroom-monitor/internal/model"
)

// CleanroomStore 洁净区存储。
type CleanroomStore interface {
	Create(ctx context.Context, c *model.Cleanroom) error
	GetByID(ctx context.Context, id int64) (*model.Cleanroom, error)
	GetByCode(ctx context.Context, code string) (*model.Cleanroom, error)
	List(ctx context.Context, limit, offset int) ([]*model.Cleanroom, error)
	Update(ctx context.Context, c *model.Cleanroom) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int, error)
}

type SQLCleanroomStore struct {
	db *DB
}

func NewCleanroomStore(db *DB) CleanroomStore {
	return &SQLCleanroomStore{db: db}
}

const cleanroomCols = "id, name, code, grade, area_sqm, status, created_at"

func scanCleanroom(row *sql.Row) (*model.Cleanroom, error) {
	var c model.Cleanroom
	var createdAt time.Time
	if err := row.Scan(&c.ID, &c.Name, &c.Code, &c.Grade, &c.AreaSqm, &c.Status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	c.CreatedAt = createdAt
	return &c, nil
}

func (s *SQLCleanroomStore) Create(ctx context.Context, c *model.Cleanroom) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO cleanrooms (name, code, grade, area_sqm, status, created_at) VALUES (?,?,?,?,?,?)`,
		c.Name, c.Code, c.Grade, c.AreaSqm, c.Status, c.CreatedAt)
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
	c.ID = id
	return nil
}

func (s *SQLCleanroomStore) GetByID(ctx context.Context, id int64) (*model.Cleanroom, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+cleanroomCols+" FROM cleanrooms WHERE id = ?", id)
	return scanCleanroom(row)
}

func (s *SQLCleanroomStore) GetByCode(ctx context.Context, code string) (*model.Cleanroom, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+cleanroomCols+" FROM cleanrooms WHERE code = ?", code)
	return scanCleanroom(row)
}

func (s *SQLCleanroomStore) List(ctx context.Context, limit, offset int) ([]*model.Cleanroom, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+cleanroomCols+" FROM cleanrooms ORDER BY id LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Cleanroom
	for rows.Next() {
		var c model.Cleanroom
		var createdAt time.Time
		if err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.Grade, &c.AreaSqm, &c.Status, &createdAt); err != nil {
			return nil, err
		}
		c.CreatedAt = createdAt
		out = append(out, &c)
	}
	return out, rows.Err()
}

func (s *SQLCleanroomStore) Update(ctx context.Context, c *model.Cleanroom) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE cleanrooms SET name=?, code=?, grade=?, area_sqm=? WHERE id=?`,
		c.Name, c.Code, c.Grade, c.AreaSqm, c.ID)
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

func (s *SQLCleanroomStore) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM cleanrooms WHERE id = ?", id)
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

func (s *SQLCleanroomStore) Count(ctx context.Context) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cleanrooms").Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
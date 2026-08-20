package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"cleanroom-monitor/internal/model"
)

// RoomStore 房间存储。
type RoomStore interface {
	Create(ctx context.Context, r *model.Room) error
	GetByID(ctx context.Context, id int64) (*model.Room, error)
	GetByCode(ctx context.Context, code string) (*model.Room, error)
	ListByCleanroom(ctx context.Context, cleanroomID int64) ([]*model.Room, error)
	List(ctx context.Context, limit, offset int) ([]*model.Room, error)
	Update(ctx context.Context, r *model.Room) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	CountByCleanroom(ctx context.Context, cleanroomID int64) (int, error)
}

type SQLRoomStore struct {
	db *DB
}

func NewRoomStore(db *DB) RoomStore {
	return &SQLRoomStore{db: db}
}

const roomCols = "id, cleanroom_id, name, code, kind, area_sqm, target_pressure, pressure_tolerance, target_temp, target_humidity, status, created_at"

func scanRoom(row interface{ Scan(...any) error }) (*model.Room, error) {
	var r model.Room
	var createdAt time.Time
	if err := row.Scan(&r.ID, &r.CleanroomID, &r.Name, &r.Code, &r.Kind, &r.AreaSqm,
		&r.TargetPressure, &r.PressureTolerance, &r.TargetTemp, &r.TargetHumidity,
		&r.Status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	r.CreatedAt = createdAt
	return &r, nil
}

func (s *SQLRoomStore) Create(ctx context.Context, r *model.Room) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO rooms (cleanroom_id, name, code, kind, area_sqm, target_pressure, pressure_tolerance, target_temp, target_humidity, status, created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		r.CleanroomID, r.Name, r.Code, r.Kind, r.AreaSqm, r.TargetPressure, r.PressureTolerance,
		r.TargetTemp, r.TargetHumidity, r.Status, r.CreatedAt)
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

func (s *SQLRoomStore) GetByID(ctx context.Context, id int64) (*model.Room, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+roomCols+" FROM rooms WHERE id = ?", id)
	return scanRoom(row)
}

func (s *SQLRoomStore) GetByCode(ctx context.Context, code string) (*model.Room, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+roomCols+" FROM rooms WHERE code = ?", code)
	return scanRoom(row)
}

func (s *SQLRoomStore) ListByCleanroom(ctx context.Context, cleanroomID int64) ([]*model.Room, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+roomCols+" FROM rooms WHERE cleanroom_id = ? ORDER BY id", cleanroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Room
	for rows.Next() {
		r, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQLRoomStore) List(ctx context.Context, limit, offset int) ([]*model.Room, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+roomCols+" FROM rooms ORDER BY id LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Room
	for rows.Next() {
		r, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQLRoomStore) Update(ctx context.Context, r *model.Room) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE rooms SET name=?, code=?, kind=?, area_sqm=?, target_pressure=?, pressure_tolerance=?, target_temp=?, target_humidity=? WHERE id=?`,
		r.Name, r.Code, r.Kind, r.AreaSqm, r.TargetPressure, r.PressureTolerance, r.TargetTemp, r.TargetHumidity, r.ID)
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

func (s *SQLRoomStore) UpdateStatus(ctx context.Context, id int64, status string) error {
	res, err := s.db.ExecContext(ctx, "UPDATE rooms SET status=? WHERE id=?", status, id)
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

func (s *SQLRoomStore) CountByCleanroom(ctx context.Context, cleanroomID int64) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM rooms WHERE cleanroom_id=?", cleanroomID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
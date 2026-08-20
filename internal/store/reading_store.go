package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"cleanroom-monitor/internal/model"
)

// ReadingQuery 读数查询条件。
type ReadingQuery struct {
	RoomID    int64
	PointID   int64
	ParamType string
	From      time.Time
	To        time.Time
	Limit     int
}

// ReadingStore 读数存储。
type ReadingStore interface {
	Insert(ctx context.Context, r *model.Reading) error
	InsertBatch(ctx context.Context, rs []*model.Reading) error
	GetByID(ctx context.Context, id int64) (*model.Reading, error)
	Query(ctx context.Context, q ReadingQuery) ([]*model.Reading, error)
	QueryValues(ctx context.Context, pointID int64, paramType string, from, to time.Time) ([]float64, error)
	QueryValuesWithTime(ctx context.Context, pointID int64, paramType string, from, to time.Time) ([]model.Reading, error)
	ExistsDup(ctx context.Context, pointID int64, measuredAt time.Time) (bool, error)
	Stats(ctx context.Context, paramType string, from, to time.Time) (*model.ReadingStats, error)
	LatestByPoint(ctx context.Context, pointID int64) (*model.Reading, error)
	CountRange(ctx context.Context, roomID int64, from, to time.Time) (int, error)
}

type SQLReadingStore struct {
	db *DB
}

func NewReadingStore(db *DB) ReadingStore {
	return &SQLReadingStore{db: db}
}

const readingCols = "id, point_id, sensor_id, param_type, value, measured_at, raw, created_at"

func scanReading(row interface{ Scan(...any) error }) (*model.Reading, error) {
	var r model.Reading
	var measured, createdAt time.Time
	var raw sql.NullString
	if err := row.Scan(&r.ID, &r.PointID, &r.SensorID, &r.ParamType, &r.Value, &measured, &raw, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	r.MeasuredAt = measured
	r.Raw = raw.String
	r.CreatedAt = createdAt
	return &r, nil
}

func (s *SQLReadingStore) Insert(ctx context.Context, r *model.Reading) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO readings (point_id, sensor_id, param_type, value, measured_at, raw, created_at) VALUES (?,?,?,?,?,?,?)`,
		r.PointID, r.SensorID, r.ParamType, r.Value, r.MeasuredAt, r.Raw, r.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = id
	return nil
}

func (s *SQLReadingStore) InsertBatch(ctx context.Context, rs []*model.Reading) error {
	return s.db.WithTx(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO readings (point_id, sensor_id, param_type, value, measured_at, raw, created_at) VALUES (?,?,?,?,?,?,?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, r := range rs {
			if _, err := stmt.ExecContext(ctx, r.PointID, r.SensorID, r.ParamType, r.Value, r.MeasuredAt, r.Raw, r.CreatedAt); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SQLReadingStore) GetByID(ctx context.Context, id int64) (*model.Reading, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+readingCols+" FROM readings WHERE id = ?", id)
	return scanReading(row)
}

func (s *SQLReadingStore) Query(ctx context.Context, q ReadingQuery) ([]*model.Reading, error) {
	qry := "SELECT " + readingCols + " FROM readings WHERE 1=1"
	args := []any{}
	if q.PointID > 0 {
		qry += " AND point_id = ?"
		args = append(args, q.PointID)
	}
	if q.ParamType != "" {
		qry += " AND param_type = ?"
		args = append(args, q.ParamType)
	}
	if !q.From.IsZero() {
		qry += " AND measured_at >= ?"
		args = append(args, q.From)
	}
	if !q.To.IsZero() {
		qry += " AND measured_at <= ?"
		args = append(args, q.To)
	}
	qry += " ORDER BY measured_at DESC"
	if q.Limit <= 0 || q.Limit > 1000 {
		q.Limit = 200
	}
	qry += " LIMIT ?"
	args = append(args, q.Limit)
	rows, err := s.db.QueryContext(ctx, qry, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Reading
	for rows.Next() {
		r, err := scanReading(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQLReadingStore) QueryValues(ctx context.Context, pointID int64, paramType string, from, to time.Time) ([]float64, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT value FROM readings WHERE point_id=? AND param_type=? AND measured_at>=? AND measured_at<=? ORDER BY measured_at`,
		pointID, paramType, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []float64
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *SQLReadingStore) QueryValuesWithTime(ctx context.Context, pointID int64, paramType string, from, to time.Time) ([]model.Reading, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+readingCols+` FROM readings WHERE point_id=? AND param_type=? AND measured_at>=? AND measured_at<=? ORDER BY measured_at`,
		pointID, paramType, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Reading
	for rows.Next() {
		r, err := scanReading(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, rows.Err()
}

func (s *SQLReadingStore) ExistsDup(ctx context.Context, pointID int64, measuredAt time.Time) (bool, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM readings WHERE point_id=? AND measured_at=?", pointID, measuredAt).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *SQLReadingStore) Stats(ctx context.Context, paramType string, from, to time.Time) (*model.ReadingStats, error) {
	st := &model.ReadingStats{ParamType: paramType, Unit: model.ParamUnits[paramType]}
	var minV, maxV, avgV float64
	var count, oor int
	row := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*), MIN(value), MAX(value), AVG(value) FROM readings WHERE param_type=? AND measured_at>=? AND measured_at<=?`,
		paramType, from, to)
	if err := row.Scan(&count, &minV, &maxV, &avgV); err != nil {
		if errors.Is(err, sql.ErrNoRows) || count == 0 {
			return st, nil
		}
		return nil, err
	}
	st.Count = count
	st.Min = minV
	st.Max = maxV
	st.Avg = avgV
	if count == 0 {
		return st, nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT value FROM readings WHERE param_type=? AND measured_at>=? AND measured_at<=?`, paramType, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []*model.MonitoringPoint
	// 该参数所有点位的阈值下界上界
	// 简化：占位，调用方负责 OOR 判定
	_ = points
	var vals []float64
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		vals = append(vals, v)
	}
	_ = vals
	_ = oor
	return st, rows.Err()
}

func (s *SQLReadingStore) LatestByPoint(ctx context.Context, pointID int64) (*model.Reading, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+readingCols+" FROM readings WHERE point_id=? ORDER BY measured_at DESC LIMIT 1", pointID)
	return scanReading(row)
}

func (s *SQLReadingStore) CountRange(ctx context.Context, roomID int64, from, to time.Time) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM readings r JOIN monitoring_points p ON r.point_id=p.id WHERE p.room_id=? AND r.measured_at>=? AND r.measured_at<=?`,
		roomID, from, to).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"seat-booking-backend/internal/config"
	"seat-booking-backend/internal/model"
)

var (
	ErrSeatNotFound         = errors.New("seat not found")
	ErrSeatNotBookable      = errors.New("seat is not bookable")
	ErrReservationConflict  = errors.New("reservation time conflict")
	ErrReservationNotFound  = errors.New("reservation not found")
	ErrReservationNotActive = errors.New("reservation is not active")
)

type Store struct {
	db       *sql.DB
	location *time.Location
}

func New(dbPath string, layout config.LayoutConfig) (*Store, error) {
	if err := ensureDBDir(dbPath); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	// SQLite is a single-file DB. For this MVP, serializing write access keeps behavior simple and predictable.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	repo := &Store{
		db:       db,
		location: mustLoadLocation("Asia/Taipei"),
	}

	if err := repo.configureSQLite(); err != nil {
		db.Close()
		return nil, err
	}
	if err := repo.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if err := repo.seedSeatsIfEmpty(layout); err != nil {
		db.Close()
		return nil, err
	}

	return repo, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func ensureDBDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create db dir: %w", err)
	}
	return nil
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}
	return loc
}

func (s *Store) configureSQLite() error {
	stmts := []string{
		`PRAGMA foreign_keys = ON;`,
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA busy_timeout = 5000;`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("configure sqlite with %q: %w", stmt, err)
		}
	}
	return nil
}

func (s *Store) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS seats (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	seat_code TEXT NOT NULL UNIQUE,
	zone_code TEXT NOT NULL,
	zone_name TEXT NOT NULL,
	seat_type TEXT NOT NULL CHECK (seat_type IN ('fixed', 'flexible')),
	fixed_owner_name TEXT NOT NULL DEFAULT '',
	is_active INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
	created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS reservations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	seat_id INTEGER NOT NULL,
	booker_name TEXT NOT NULL,
	start_ts INTEGER NOT NULL,
	end_ts INTEGER NOT NULL,
	status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
	note TEXT NOT NULL DEFAULT '',
	created_at INTEGER NOT NULL,
	cancelled_at INTEGER NULL,
	FOREIGN KEY (seat_id) REFERENCES seats(id)
);

CREATE INDEX IF NOT EXISTS idx_seats_zone_code ON seats(zone_code);
CREATE INDEX IF NOT EXISTS idx_seats_type ON seats(seat_type);
CREATE INDEX IF NOT EXISTS idx_reservations_seat_status_time
	ON reservations(seat_id, status, start_ts, end_ts);
CREATE INDEX IF NOT EXISTS idx_reservations_status_created_at
	ON reservations(status, created_at);
`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	return nil
}

func (s *Store) seedSeatsIfEmpty(layout config.LayoutConfig) error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM seats`).Scan(&count); err != nil {
		return fmt.Errorf("count seats: %w", err)
	}
	if count > 0 {
		return nil
	}

	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
INSERT INTO seats (
	seat_code, zone_code, zone_name, seat_type,
	fixed_owner_name, is_active, created_at
) VALUES (?, ?, ?, ?, ?, 1, ?)
`)
	if err != nil {
		return fmt.Errorf("prepare seed insert: %w", err)
	}
	defer stmt.Close()

	seatNumber := layout.SeatNumberStart
	now := time.Now().Unix()

	for _, zone := range layout.Zones {
		for localIndex := 1; localIndex <= zone.SeatCount; localIndex++ {
			code := layout.SeatCodePrefix + fmt.Sprintf("%0*d", layout.SeatNumberWidth, seatNumber)
			seatType := "flexible"
			fixedOwner := ""

			if localIndex <= zone.FixedCount {
				seatType = "fixed"
				fixedOwner = strings.TrimSpace(layout.FixedOwnerMap[code])
			}

			if _, err := stmt.Exec(code, zone.ZoneCode, zone.ZoneName, seatType, fixedOwner, now); err != nil {
				return fmt.Errorf("seed seat %s: %w", code, err)
			}
			seatNumber++
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seat seed tx: %w", err)
	}
	return nil
}

func (s *Store) ParseInputTime(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, fmt.Errorf("time cannot be empty")
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02 15:04",
	}

	var lastErr error
	for _, layout := range formats {
		if layout == time.RFC3339 {
			t, err := time.Parse(layout, value)
			if err == nil {
				return t, nil
			}
			lastErr = err
			continue
		}

		t, err := time.ParseInLocation(layout, value, s.location)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}

	return time.Time{}, fmt.Errorf("unsupported time format %q: %w", raw, lastErr)
}

func (s *Store) ListSeats(ctx context.Context, start, end time.Time) ([]model.Seat, error) {
	startTS := start.Unix()
	endTS := end.Unix()

	rows, err := s.db.QueryContext(ctx, `
SELECT
	s.id,
	s.seat_code,
	s.zone_code,
	s.zone_name,
	s.seat_type,
	s.fixed_owner_name,
	s.is_active,
	CASE
		WHEN s.is_active = 0 THEN 'disabled'
		WHEN s.seat_type = 'fixed' THEN 'fixed'
		WHEN EXISTS (
			SELECT 1
			FROM reservations r
			WHERE r.seat_id = s.id
			  AND r.status = 'active'
			  AND r.start_ts < ?
			  AND r.end_ts > ?
		) THEN 'booked'
		ELSE 'available'
	END AS availability,
	(
		SELECT r.booker_name
		FROM reservations r
		WHERE r.seat_id = s.id
		  AND r.status = 'active'
		  AND r.start_ts < ?
		  AND r.end_ts > ?
		ORDER BY r.created_at DESC
		LIMIT 1
	) AS current_booker,
	(
		SELECT r.id
		FROM reservations r
		WHERE r.seat_id = s.id
		  AND r.status = 'active'
		  AND r.start_ts < ?
		  AND r.end_ts > ?
		ORDER BY r.created_at DESC
		LIMIT 1
	) AS current_booking_id
FROM seats s
ORDER BY s.id ASC;
`, endTS, startTS, endTS, startTS, endTS, startTS)
	if err != nil {
		return nil, fmt.Errorf("query seats: %w", err)
	}
	defer rows.Close()

	var result []model.Seat
	for rows.Next() {
		var (
			item             model.Seat
			isActiveInt      int
			currentBooker    sql.NullString
			currentBookingID sql.NullInt64
		)
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.ZoneCode,
			&item.ZoneName,
			&item.SeatType,
			&item.FixedOwnerName,
			&isActiveInt,
			&item.Availability,
			&currentBooker,
			&currentBookingID,
		); err != nil {
			return nil, fmt.Errorf("scan seat: %w", err)
		}
		item.IsActive = isActiveInt == 1
		if currentBooker.Valid {
			item.CurrentBooker = currentBooker.String
		}
		if currentBookingID.Valid {
			value := currentBookingID.Int64
			item.CurrentBookingID = &value
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate seats: %w", err)
	}
	return result, nil
}

func (s *Store) CreateReservation(ctx context.Context, req model.CreateReservationRequest) (model.Reservation, error) {
	seatCode := strings.TrimSpace(req.SeatCode)
	bookerName := strings.TrimSpace(req.BookerName)
	note := strings.TrimSpace(req.Note)

	if seatCode == "" {
		return model.Reservation{}, fmt.Errorf("seat_code cannot be empty")
	}
	if bookerName == "" {
		return model.Reservation{}, fmt.Errorf("booker_name cannot be empty")
	}

	start, err := s.ParseInputTime(req.StartTime)
	if err != nil {
		return model.Reservation{}, fmt.Errorf("invalid start_time: %w", err)
	}
	end, err := s.ParseInputTime(req.EndTime)
	if err != nil {
		return model.Reservation{}, fmt.Errorf("invalid end_time: %w", err)
	}
	if !start.Before(end) {
		return model.Reservation{}, fmt.Errorf("end_time must be later than start_time")
	}

	startTS := start.Unix()
	endTS := end.Unix()
	nowTS := time.Now().Unix()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Reservation{}, fmt.Errorf("begin reservation tx: %w", err)
	}
	defer tx.Rollback()

	var seat struct {
		ID       int64
		Code     string
		ZoneCode string
		ZoneName string
		SeatType string
		IsActive int
	}
	err = tx.QueryRowContext(ctx, `
SELECT id, seat_code, zone_code, zone_name, seat_type, is_active
FROM seats
WHERE seat_code = ?
`, seatCode).Scan(
		&seat.ID,
		&seat.Code,
		&seat.ZoneCode,
		&seat.ZoneName,
		&seat.SeatType,
		&seat.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Reservation{}, ErrSeatNotFound
	}
	if err != nil {
		return model.Reservation{}, fmt.Errorf("query seat: %w", err)
	}

	if seat.IsActive != 1 || seat.SeatType != "flexible" {
		return model.Reservation{}, ErrSeatNotBookable
	}

	var conflictCount int
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(1)
FROM reservations
WHERE seat_id = ?
  AND status = 'active'
  AND start_ts < ?
  AND end_ts > ?
`, seat.ID, endTS, startTS).Scan(&conflictCount); err != nil {
		return model.Reservation{}, fmt.Errorf("check reservation conflict: %w", err)
	}
	if conflictCount > 0 {
		return model.Reservation{}, ErrReservationConflict
	}

	result, err := tx.ExecContext(ctx, `
INSERT INTO reservations (
	seat_id, booker_name, start_ts, end_ts,
	status, note, created_at
) VALUES (?, ?, ?, ?, 'active', ?, ?)
`, seat.ID, bookerName, startTS, endTS, note, nowTS)
	if err != nil {
		return model.Reservation{}, fmt.Errorf("insert reservation: %w", err)
	}

	reservationID, err := result.LastInsertId()
	if err != nil {
		return model.Reservation{}, fmt.Errorf("get reservation id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.Reservation{}, fmt.Errorf("commit reservation tx: %w", err)
	}

	return model.Reservation{
		ID:         reservationID,
		SeatID:     seat.ID,
		SeatCode:   seat.Code,
		ZoneCode:   seat.ZoneCode,
		ZoneName:   seat.ZoneName,
		BookerName: bookerName,
		StartTime:  start.In(s.location).Format(time.RFC3339),
		EndTime:    end.In(s.location).Format(time.RFC3339),
		Status:     "active",
		Note:       note,
		CreatedAt:  time.Unix(nowTS, 0).In(s.location).Format(time.RFC3339),
	}, nil
}

func (s *Store) CancelReservation(ctx context.Context, id int64) (model.Reservation, error) {
	if id <= 0 {
		return model.Reservation{}, ErrReservationNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Reservation{}, fmt.Errorf("begin cancel reservation tx: %w", err)
	}
	defer tx.Rollback()

	reservation, err := s.getReservationByIDTx(ctx, tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Reservation{}, ErrReservationNotFound
	}
	if err != nil {
		return model.Reservation{}, err
	}
	if reservation.Status != "active" {
		return model.Reservation{}, ErrReservationNotActive
	}

	cancelledAt := time.Now().Unix()
	if _, err := tx.ExecContext(ctx, `
UPDATE reservations
SET status = 'cancelled',
	cancelled_at = ?
WHERE id = ?
`, cancelledAt, id); err != nil {
		return model.Reservation{}, fmt.Errorf("cancel reservation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.Reservation{}, fmt.Errorf("commit cancel reservation tx: %w", err)
	}

	reservation.Status = "cancelled"
	reservation.CancelledAt = time.Unix(cancelledAt, 0).In(s.location).Format(time.RFC3339)
	return reservation, nil
}

func (s *Store) getReservationByIDTx(ctx context.Context, tx *sql.Tx, id int64) (model.Reservation, error) {
	var (
		item        model.Reservation
		startTS     int64
		endTS       int64
		createdTS   int64
		cancelledTS sql.NullInt64
	)
	err := tx.QueryRowContext(ctx, `
SELECT
	r.id,
	r.seat_id,
	s.seat_code,
	s.zone_code,
	s.zone_name,
	r.booker_name,
	r.start_ts,
	r.end_ts,
	r.status,
	r.note,
	r.created_at,
	r.cancelled_at
FROM reservations r
JOIN seats s ON s.id = r.seat_id
WHERE r.id = ?
`, id).Scan(
		&item.ID,
		&item.SeatID,
		&item.SeatCode,
		&item.ZoneCode,
		&item.ZoneName,
		&item.BookerName,
		&startTS,
		&endTS,
		&item.Status,
		&item.Note,
		&createdTS,
		&cancelledTS,
	)
	if err != nil {
		return model.Reservation{}, fmt.Errorf("get reservation by id: %w", err)
	}

	item.StartTime = time.Unix(startTS, 0).In(s.location).Format(time.RFC3339)
	item.EndTime = time.Unix(endTS, 0).In(s.location).Format(time.RFC3339)
	item.CreatedAt = time.Unix(createdTS, 0).In(s.location).Format(time.RFC3339)
	if cancelledTS.Valid {
		item.CancelledAt = time.Unix(cancelledTS.Int64, 0).In(s.location).Format(time.RFC3339)
	}
	return item, nil
}

type ReservationQuery struct {
	Status    string
	SeatCode  string
	StartTime *time.Time
	EndTime   *time.Time
}

func (s *Store) ListReservations(ctx context.Context, query ReservationQuery) ([]model.Reservation, error) {
	clauses := []string{"1 = 1"}
	args := []any{}

	if query.Status != "" && query.Status != "all" {
		clauses = append(clauses, "r.status = ?")
		args = append(args, query.Status)
	}
	if strings.TrimSpace(query.SeatCode) != "" {
		clauses = append(clauses, "s.seat_code = ?")
		args = append(args, strings.TrimSpace(query.SeatCode))
	}
	if query.StartTime != nil && query.EndTime != nil {
		clauses = append(clauses, "r.start_ts < ? AND r.end_ts > ?")
		args = append(args, query.EndTime.Unix(), query.StartTime.Unix())
	}

	sqlText := fmt.Sprintf(`
SELECT
	r.id,
	r.seat_id,
	s.seat_code,
	s.zone_code,
	s.zone_name,
	r.booker_name,
	r.start_ts,
	r.end_ts,
	r.status,
	r.note,
	r.created_at,
	r.cancelled_at
FROM reservations r
JOIN seats s ON s.id = r.seat_id
WHERE %s
ORDER BY r.created_at DESC, r.id DESC
`, strings.Join(clauses, " AND "))

	rows, err := s.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("query reservations: %w", err)
	}
	defer rows.Close()

	var result []model.Reservation
	for rows.Next() {
		var (
			item        model.Reservation
			startTS     int64
			endTS       int64
			createdTS   int64
			cancelledTS sql.NullInt64
		)
		if err := rows.Scan(
			&item.ID,
			&item.SeatID,
			&item.SeatCode,
			&item.ZoneCode,
			&item.ZoneName,
			&item.BookerName,
			&startTS,
			&endTS,
			&item.Status,
			&item.Note,
			&createdTS,
			&cancelledTS,
		); err != nil {
			return nil, fmt.Errorf("scan reservation: %w", err)
		}

		item.StartTime = time.Unix(startTS, 0).In(s.location).Format(time.RFC3339)
		item.EndTime = time.Unix(endTS, 0).In(s.location).Format(time.RFC3339)
		item.CreatedAt = time.Unix(createdTS, 0).In(s.location).Format(time.RFC3339)
		if cancelledTS.Valid {
			item.CancelledAt = time.Unix(cancelledTS.Int64, 0).In(s.location).Format(time.RFC3339)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reservations: %w", err)
	}
	return result, nil
}

func ParseReservationID(raw string) (int64, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, fmt.Errorf("reservation id cannot be empty")
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid reservation id")
	}
	return id, nil
}

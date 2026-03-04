package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
	"schedulehelper/backend/models"
)

// DB wraps sql.DB with helper methods
type DB struct {
	*sql.DB
}

// Open opens (or creates) the SQLite database and applies migrations
func Open(dsn string) (*DB, error) {
	sqldb, err := sql.Open("sqlite", dsn+"?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	sqldb.SetMaxOpenConns(1)
	d := &DB{sqldb}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return d, nil
}

func (d *DB) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	name        TEXT    NOT NULL UNIQUE,
	emoji       TEXT    NOT NULL DEFAULT '😊',
	color       TEXT    NOT NULL DEFAULT '#4A90D9',
	theme       TEXT    NOT NULL DEFAULT 'system',
	timezone    TEXT    NOT NULL DEFAULT 'UTC',
	auto_tz     INTEGER NOT NULL DEFAULT 1,
	sort_order  INTEGER NOT NULL DEFAULT 0,
	password_hash TEXT,
	gcal_token    TEXT,
	created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS slots (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	start_time  DATETIME NOT NULL,
	end_time    DATETIME NOT NULL,
	comment     TEXT    NOT NULL DEFAULT '',
	created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS events (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	title       TEXT    NOT NULL,
	description TEXT    NOT NULL DEFAULT '',
	icon        TEXT    NOT NULL DEFAULT '',
	color       TEXT    NOT NULL DEFAULT '#D4AF37',
	start_time  DATETIME NOT NULL,
	end_time    DATETIME NOT NULL,
	created_by  INTEGER NOT NULL REFERENCES users(id) ON DELETE SET NULL,
	created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS event_participants (
	event_id    INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
	user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	PRIMARY KEY (event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_slots_user_id ON slots(user_id);
CREATE INDEX IF NOT EXISTS idx_slots_time ON slots(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_events_time ON events(start_time, end_time);
`
	_, err := d.Exec(schema)
	return err
}

// ---- USER OPERATIONS ----

func (d *DB) GetAllUsers() ([]models.User, error) {
	rows, err := d.Query(`
		SELECT id, name, emoji, color, theme, timezone, auto_tz, sort_order, created_at
		FROM users ORDER BY sort_order ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var u models.User
		var autoTZ int
		if err := rows.Scan(&u.ID, &u.Name, &u.Emoji, &u.Color, &u.Theme, &u.Timezone, &autoTZ, &u.SortOrder, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.AutoTZ = autoTZ == 1
		users = append(users, u)
	}
	if users == nil {
		users = []models.User{}
	}
	return users, rows.Err()
}

func (d *DB) GetUserByID(id int64) (*models.User, error) {
	row := d.QueryRow(`
		SELECT id, name, emoji, color, theme, timezone, auto_tz, sort_order, created_at
		FROM users WHERE id = ?
	`, id)
	var u models.User
	var autoTZ int
	if err := row.Scan(&u.ID, &u.Name, &u.Emoji, &u.Color, &u.Theme, &u.Timezone, &autoTZ, &u.SortOrder, &u.CreatedAt); err != nil {
		return nil, err
	}
	u.AutoTZ = autoTZ == 1
	return &u, nil
}

func (d *DB) GetUserByName(name string) (*models.User, error) {
	row := d.QueryRow(`
		SELECT id, name, emoji, color, theme, timezone, auto_tz, sort_order, created_at
		FROM users WHERE name = ?
	`, name)
	var u models.User
	var autoTZ int
	if err := row.Scan(&u.ID, &u.Name, &u.Emoji, &u.Color, &u.Theme, &u.Timezone, &autoTZ, &u.SortOrder, &u.CreatedAt); err != nil {
		return nil, err
	}
	u.AutoTZ = autoTZ == 1
	return &u, nil
}

func (d *DB) CreateUser(req models.RegisterRequest) (*models.User, error) {
	tz := req.Timezone
	if tz == "" {
		tz = "UTC"
	}
	res, err := d.Exec(`
		INSERT INTO users (name, emoji, color, timezone, auto_tz, sort_order)
		VALUES (?, ?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM users))
	`, req.Name, req.Emoji, req.Color, tz, boolToInt(req.AutoTZ))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return d.GetUserByID(id)
}

func (d *DB) UpdateUser(id int64, req models.UpdateProfileRequest) (*models.User, error) {
	_, err := d.Exec(`
		UPDATE users SET name=?, emoji=?, color=?, theme=?, timezone=?, auto_tz=?
		WHERE id=?
	`, req.Name, req.Emoji, req.Color, req.Theme, req.Timezone, boolToInt(req.AutoTZ), id)
	if err != nil {
		return nil, err
	}
	return d.GetUserByID(id)
}

func (d *DB) DeleteUser(id int64) error {
	_, err := d.Exec(`DELETE FROM users WHERE id=?`, id)
	return err
}

func (d *DB) UpdateSortOrder(order []int64) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	for i, uid := range order {
		if _, err = tx.Exec(`UPDATE users SET sort_order=? WHERE id=?`, i, uid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ---- SLOT OPERATIONS ----

func (d *DB) GetSlots(userID int64, from, to time.Time) ([]models.Slot, error) {
	query := `SELECT id, user_id, start_time, end_time, comment, created_at, updated_at FROM slots WHERE 1=1`
	args := []interface{}{}
	if userID > 0 {
		query += " AND user_id=?"
		args = append(args, userID)
	}
	if !from.IsZero() {
		query += " AND end_time >= ?"
		args = append(args, from)
	}
	if !to.IsZero() {
		query += " AND start_time <= ?"
		args = append(args, to)
	}
	query += " ORDER BY start_time ASC"

	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slots []models.Slot
	for rows.Next() {
		var s models.Slot
		if err := rows.Scan(&s.ID, &s.UserID, &s.StartTime, &s.EndTime, &s.Comment, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	if slots == nil {
		slots = []models.Slot{}
	}
	return slots, rows.Err()
}

func (d *DB) CreateSlot(userID int64, startTime, endTime time.Time, comment string) (*models.Slot, error) {
	res, err := d.Exec(`
		INSERT INTO slots (user_id, start_time, end_time, comment)
		VALUES (?, ?, ?, ?)
	`, userID, startTime, endTime, comment)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return d.GetSlotByID(id)
}

func (d *DB) GetSlotByID(id int64) (*models.Slot, error) {
	row := d.QueryRow(`
		SELECT id, user_id, start_time, end_time, comment, created_at, updated_at
		FROM slots WHERE id=?
	`, id)
	var s models.Slot
	if err := row.Scan(&s.ID, &s.UserID, &s.StartTime, &s.EndTime, &s.Comment, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

func (d *DB) UpdateSlot(id int64, startTime, endTime time.Time, comment string) (*models.Slot, error) {
	_, err := d.Exec(`
		UPDATE slots SET start_time=?, end_time=?, comment=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?
	`, startTime, endTime, comment, id)
	if err != nil {
		return nil, err
	}
	return d.GetSlotByID(id)
}

func (d *DB) DeleteSlot(id int64) error {
	_, err := d.Exec(`DELETE FROM slots WHERE id=?`, id)
	return err
}

// ---- EVENT OPERATIONS ----

func (d *DB) GetEvents(from, to time.Time) ([]models.Event, error) {
	query := `SELECT id, title, description, icon, color, start_time, end_time, created_by, created_at, updated_at FROM events WHERE 1=1`
	args := []interface{}{}
	if !from.IsZero() {
		query += " AND end_time >= ?"
		args = append(args, from)
	}
	if !to.IsZero() {
		query += " AND start_time <= ?"
		args = append(args, to)
	}
	query += " ORDER BY start_time ASC"
	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.Icon, &e.Color, &e.StartTime, &e.EndTime, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if events == nil {
		events = []models.Event{}
	}
	return events, rows.Err()
}

func (d *DB) GetEventByID(id int64) (*models.Event, error) {
	row := d.QueryRow(`
		SELECT id, title, description, icon, color, start_time, end_time, created_by, created_at, updated_at
		FROM events WHERE id=?
	`, id)
	var e models.Event
	if err := row.Scan(&e.ID, &e.Title, &e.Description, &e.Icon, &e.Color, &e.StartTime, &e.EndTime, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, err
	}
	return &e, nil
}

func (d *DB) CreateEvent(e models.Event) (*models.Event, error) {
	res, err := d.Exec(`
		INSERT INTO events (title, description, icon, color, start_time, end_time, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, e.Title, e.Description, e.Icon, e.Color, e.StartTime, e.EndTime, e.CreatedBy)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return d.GetEventByID(id)
}

func (d *DB) UpdateEvent(id int64, e models.Event) (*models.Event, error) {
	_, err := d.Exec(`
		UPDATE events SET title=?, description=?, icon=?, color=?, start_time=?, end_time=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?
	`, e.Title, e.Description, e.Icon, e.Color, e.StartTime, e.EndTime, id)
	if err != nil {
		return nil, err
	}
	return d.GetEventByID(id)
}

func (d *DB) DeleteEvent(id int64) error {
	_, err := d.Exec(`DELETE FROM events WHERE id=?`, id)
	return err
}

// ---- CLEANUP ----

func (d *DB) CleanupOldData() {
	cutoff := time.Now().Add(-6 * 30 * 24 * time.Hour)
	res, err := d.Exec(`DELETE FROM slots WHERE end_time < ?`, cutoff)
	if err != nil {
		log.Printf("cleanup slots: %v", err)
	} else {
		n, _ := res.RowsAffected()
		log.Printf("cleanup: deleted %d old slots", n)
	}
	res, err = d.Exec(`DELETE FROM events WHERE end_time < ?`, cutoff)
	if err != nil {
		log.Printf("cleanup events: %v", err)
	} else {
		n, _ := res.RowsAffected()
		log.Printf("cleanup: deleted %d old events", n)
	}
}

// ---- HELPERS ----

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

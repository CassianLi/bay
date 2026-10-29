package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CassianLi/bay/internal/model"
	_ "modernc.org/sqlite"
)

var (
	ErrNotFound  = errors.New("project not found")
	ErrAmbiguous = errors.New("ambiguous project name")
)

type Store struct {
	db *sql.DB
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS projects (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  root TEXT NOT NULL,
  path TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL,
  purpose_known INTEGER NOT NULL DEFAULT 0,
  summary TEXT NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  signals TEXT NOT NULL DEFAULT '',
  user_set_status INTEGER NOT NULL DEFAULT 0,
  missing INTEGER NOT NULL DEFAULT 0,
  last_seen_at TEXT NOT NULL,
  last_moved_at TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(root, name)
);
`

const projectSelectCols = `id, name, root, path, status, purpose_known, summary, note, signals,
       user_set_status, missing, last_seen_at, last_moved_at, created_at, updated_at`

func Open(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pragma foreign_keys: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := ensureColumn(db, "projects", "note", "TEXT NOT NULL DEFAULT ''"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate note: %w", err)
	}
	return &Store{db: db}, nil
}

func ensureColumn(db *sql.DB, table, column, decl string) error {
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			cid       int
			name      string
			ctype     string
			notnull   int
			dfltValue sql.NullString
			pk        int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + decl)
	return err
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) UpsertScan(p model.Project) error {
	now := time.Now().UTC().Format(time.RFC3339)
	lastSeen := formatTime(p.LastSeenAt)
	created := formatTime(p.CreatedAt)
	if created == "" {
		created = now
	}
	updated := formatTime(p.UpdatedAt)
	if updated == "" {
		updated = now
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		id            int64
		userSetStatus int
		status        string
	)
	err = tx.QueryRow(
		`SELECT id, user_set_status, status FROM projects WHERE root = ? AND name = ?`,
		p.Root, p.Name,
	).Scan(&id, &userSetStatus, &status)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err = tx.Exec(`
INSERT INTO projects (
  name, root, path, status, purpose_known, summary, note, signals,
  user_set_status, missing, last_seen_at, last_moved_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?)`,
			p.Name, p.Root, p.Path, string(p.Status),
			boolToInt(p.PurposeKnown), p.Summary, p.Note, p.Signals,
			boolToInt(p.UserSetStatus), lastSeen, formatTime(p.LastMovedAt),
			created, updated,
		)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		newStatus := status
		if userSetStatus == 0 {
			newStatus = string(p.Status)
		}
		_, err = tx.Exec(`
UPDATE projects SET
  path = ?, signals = ?, last_seen_at = ?, missing = 0, status = ?, updated_at = ?
WHERE id = ?`,
			p.Path, p.Signals, lastSeen, newStatus, now, id,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) GetByName(name, root string) (model.Project, error) {
	if root != "" {
		row := s.db.QueryRow(`
SELECT `+projectSelectCols+`
FROM projects WHERE name = ? AND root = ?`, name, root)
		p, err := scanProject(row)
		if errors.Is(err, sql.ErrNoRows) {
			return model.Project{}, ErrNotFound
		}
		return p, err
	}

	rows, err := s.db.Query(`
SELECT `+projectSelectCols+`
FROM projects WHERE name = ?`, name)
	if err != nil {
		return model.Project{}, err
	}
	defer rows.Close()

	var found []model.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return model.Project{}, err
		}
		found = append(found, p)
	}
	if err := rows.Err(); err != nil {
		return model.Project{}, err
	}
	switch len(found) {
	case 0:
		return model.Project{}, ErrNotFound
	case 1:
		return found[0], nil
	default:
		return model.Project{}, ErrAmbiguous
	}
}

func (s *Store) List(f model.ListFilter) ([]model.Project, error) {
	var (
		conds []string
		args  []any
	)
	if f.Status != nil {
		conds = append(conds, "status = ?")
		args = append(args, string(*f.Status))
	}
	if f.UnknownOnly {
		conds = append(conds, "purpose_known = 0")
	}
	if f.Root != "" {
		conds = append(conds, "root = ?")
		args = append(args, f.Root)
	}
	if f.Query != "" {
		conds = append(conds, "(name LIKE ? OR summary LIKE ?)")
		q := "%" + f.Query + "%"
		args = append(args, q, q)
	}

	query := `
SELECT ` + projectSelectCols + `
FROM projects`
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY name, root"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) UpdateMeta(p model.Project) error {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(`
UPDATE projects SET
  status = ?, summary = ?, note = ?, purpose_known = ?, user_set_status = ?,
  path = ?, last_moved_at = ?, updated_at = ?
WHERE root = ? AND name = ?`,
		string(p.Status), p.Summary, p.Note, boolToInt(p.PurposeKnown), boolToInt(p.UserSetStatus),
		p.Path, formatTime(p.LastMovedAt), now,
		p.Root, p.Name,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) MarkMissingExcept(root string, seenNames []string) (int, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	if len(seenNames) == 0 {
		res, err := s.db.Exec(`
UPDATE projects SET missing = 1, updated_at = ?
WHERE root = ? AND missing = 0`, now, root)
		if err != nil {
			return 0, err
		}
		n, err := res.RowsAffected()
		return int(n), err
	}

	placeholders := make([]string, len(seenNames))
	args := make([]any, 0, len(seenNames)+2)
	args = append(args, now, root)
	for i, name := range seenNames {
		placeholders[i] = "?"
		args = append(args, name)
	}
	q := fmt.Sprintf(`
UPDATE projects SET missing = 1, updated_at = ?
WHERE root = ? AND missing = 0 AND name NOT IN (%s)`, strings.Join(placeholders, ","))
	res, err := s.db.Exec(q, args...)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}

type scannable interface {
	Scan(dest ...any) error
}

func scanProject(row scannable) (model.Project, error) {
	var (
		p                                     model.Project
		status                                string
		purposeKnown, userSetStatus, missing  int
		lastSeen, lastMoved, created, updated string
	)
	err := row.Scan(
		&p.ID, &p.Name, &p.Root, &p.Path, &status,
		&purposeKnown, &p.Summary, &p.Note, &p.Signals,
		&userSetStatus, &missing,
		&lastSeen, &lastMoved, &created, &updated,
	)
	if err != nil {
		return model.Project{}, err
	}
	p.Status = model.Status(status)
	p.PurposeKnown = purposeKnown != 0
	p.UserSetStatus = userSetStatus != 0
	p.Missing = missing != 0
	p.LastSeenAt, err = parseTime(lastSeen)
	if err != nil {
		return model.Project{}, err
	}
	p.LastMovedAt, err = parseTime(lastMoved)
	if err != nil {
		return model.Project{}, err
	}
	p.CreatedAt, err = parseTime(created)
	if err != nil {
		return model.Project{}, err
	}
	p.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return model.Project{}, err
	}
	return p, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, s)
}

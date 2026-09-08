package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/KyleBing/portal-go/migrations"
)

var migrationFileRE = regexp.MustCompile(`^\d{3}_.+\.sql$`)

// Migrate applies embedded numbered SQL migrations (NNN_*.sql) against diary.
// init.sql is intentionally excluded; it is only used by first-time setup.
func Migrate(sqlDB *sql.DB) error {
	if _, err := sqlDB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name VARCHAR(255) PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`); err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}

	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !migrationFileRE.MatchString(e.Name()) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		var applied int
		if err := sqlDB.QueryRow(`SELECT COUNT(1) FROM schema_migrations WHERE name = ?`, name).Scan(&applied); err != nil {
			return err
		}
		if applied > 0 {
			fmt.Printf("%-10s %s\n", "skip", name)
			continue
		}
		body, err := migrations.FS.ReadFile(name)
		if err != nil {
			return err
		}
		tx, err := sqlDB.Begin()
		if err != nil {
			return err
		}
		for _, stmt := range splitSQL(string(body)) {
			if _, err := tx.Exec(stmt); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("%s: %w\nstmt: %s", name, err, stmt)
			}
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, name); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		fmt.Printf("%-10s %s\n", "applied", name)
	}
	return nil
}

func splitSQL(sqlText string) []string {
	var out []string
	var buf strings.Builder
	for _, line := range strings.Split(sqlText, "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "--") {
			continue
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			out = append(out, strings.TrimSpace(buf.String()))
			buf.Reset()
		}
	}
	if s := strings.TrimSpace(buf.String()); s != "" {
		out = append(out, s)
	}
	return out
}

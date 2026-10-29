package billimport

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"time"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/util"
)

//go:embed default_aliases.json
var defaultAliasesJSON []byte

func openDB() (*sql.DB, error) {
	return db.Open(db.Diary)
}

func defaultAliasConfig() (AliasConfig, error) {
	var cfg AliasConfig
	if err := json.Unmarshal(defaultAliasesJSON, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// ensureAliasConfig 无配置时写入默认种子
func ensureAliasConfig(uid int64) (AliasConfig, error) {
	diary, err := openDB()
	if err != nil {
		return AliasConfig{}, err
	}
	row := diary.QueryRow(`select slots, rules from bill_alias_config where uid=?`, uid)
	var slotsRaw, rulesRaw []byte
	err = row.Scan(&slotsRaw, &rulesRaw)
	if err == nil {
		var cfg AliasConfig
		if err := json.Unmarshal(slotsRaw, &cfg.Slots); err != nil {
			return cfg, err
		}
		if err := json.Unmarshal(rulesRaw, &cfg.Rules); err != nil {
			return cfg, err
		}
		return cfg, nil
	}
	if err != sql.ErrNoRows {
		return AliasConfig{}, err
	}
	cfg, err := defaultAliasConfig()
	if err != nil {
		return cfg, err
	}
	if err := saveAliasConfig(uid, cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func saveAliasConfig(uid int64, cfg AliasConfig) error {
	diary, err := openDB()
	if err != nil {
		return err
	}
	slots, err := json.Marshal(cfg.Slots)
	if err != nil {
		return err
	}
	rules, err := json.Marshal(cfg.Rules)
	if err != nil {
		return err
	}
	now := util.NowString()
	_, err = diary.Exec(`
		INSERT INTO bill_alias_config(uid, slots, rules, date_modify) VALUES(?,?,?,?)
		ON DUPLICATE KEY UPDATE slots=VALUES(slots), rules=VALUES(rules), date_modify=VALUES(date_modify)`,
		uid, slots, rules, now)
	return err
}

type batchRow struct {
	ID            int64
	UID           int64
	Filename      string
	StoredPath    string
	PeriodStart   sql.NullTime
	PeriodEnd     sql.NullTime
	RowCount      int
	ReadyCount    int
	GuessCount    int
	ConfirmCount  int
	UnmappedCount int
	SkipCount     int
	AppliedCount  int
	Status        string
	DateInit      sql.NullTime
}

func insertBatch(uid int64, filename, storedPath string, periodStart, periodEnd *time.Time, counts Counts) (int64, error) {
	diary, err := openDB()
	if err != nil {
		return 0, err
	}
	now := util.NowString()
	var ps, pe interface{}
	if periodStart != nil {
		ps = periodStart.Format("2006-01-02 15:04:05")
	}
	if periodEnd != nil {
		pe = periodEnd.Format("2006-01-02 15:04:05")
	}
	res, err := diary.Exec(`
		INSERT INTO bill_import_batch(
			uid, filename, stored_path, period_start, period_end,
			row_count, ready_count, guess_count, confirm_count, unmapped_count, skip_count,
			applied_count, status, date_init
		) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		uid, filename, storedPath, ps, pe,
		counts.Row, counts.Ready, counts.Guess, counts.Confirm, counts.Unmapped, counts.Skip,
		0, statusUploaded, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func updateBatchCounts(batchID int64, counts Counts) error {
	diary, err := openDB()
	if err != nil {
		return err
	}
	_, err = diary.Exec(`
		UPDATE bill_import_batch SET
			row_count=?, ready_count=?, guess_count=?, confirm_count=?, unmapped_count=?, skip_count=?
		WHERE id=?`,
		counts.Row, counts.Ready, counts.Guess, counts.Confirm, counts.Unmapped, counts.Skip, batchID)
	return err
}

func updateBatchApply(batchID int64, applied int, status string) error {
	diary, err := openDB()
	if err != nil {
		return err
	}
	_, err = diary.Exec(`UPDATE bill_import_batch SET applied_count=?, status=? WHERE id=?`, applied, status, batchID)
	return err
}

func getBatch(uid, batchID int64) (*batchRow, error) {
	diary, err := openDB()
	if err != nil {
		return nil, err
	}
	b := &batchRow{}
	err = diary.QueryRow(`
		SELECT id, uid, filename, stored_path, period_start, period_end,
			row_count, ready_count, guess_count, confirm_count, unmapped_count, skip_count,
			applied_count, status, date_init
		FROM bill_import_batch WHERE id=? AND uid=?`, batchID, uid).Scan(
		&b.ID, &b.UID, &b.Filename, &b.StoredPath, &b.PeriodStart, &b.PeriodEnd,
		&b.RowCount, &b.ReadyCount, &b.GuessCount, &b.ConfirmCount, &b.UnmappedCount, &b.SkipCount,
		&b.AppliedCount, &b.Status, &b.DateInit,
	)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func listBatches(uid int64, pageNo, pageSize int) ([]map[string]interface{}, int, error) {
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	diary, err := openDB()
	if err != nil {
		return nil, 0, err
	}
	var total int
	if err := diary.QueryRow(`SELECT COUNT(*) FROM bill_import_batch WHERE uid=?`, uid).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (pageNo - 1) * pageSize
	rows, err := db.QueryMaps(diary, `
		SELECT id, filename, stored_path, period_start, period_end,
			row_count, ready_count, guess_count, confirm_count, unmapped_count, skip_count,
			applied_count, status, date_init
		FROM bill_import_batch WHERE uid=? ORDER BY id DESC LIMIT ? OFFSET ?`, uid, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func knownOrderIDs(uid int64) (map[string]bool, error) {
	diary, err := openDB()
	if err != nil {
		return nil, err
	}
	rows, err := diary.Query(`SELECT order_id FROM bill_import_order WHERE uid=?`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

func insertOrders(uid, batchID int64, lines []DayLine) error {
	diary, err := openDB()
	if err != nil {
		return err
	}
	now := util.NowString()
	tx, err := diary.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT IGNORE INTO bill_import_order(uid, order_id, batch_id, bill_date, name, amount, date_init)
		VALUES(?,?,?,?,?,?,?)`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, line := range lines {
		for _, oid := range line.OrderIDs {
			if oid == "" {
				continue
			}
			if _, err := stmt.Exec(uid, oid, batchID, line.Date, line.Name, line.Amount, now); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit()
}

func listOrders(uid, batchID int64) ([]map[string]interface{}, error) {
	diary, err := openDB()
	if err != nil {
		return nil, err
	}
	return db.QueryMaps(diary, `
		SELECT id, order_id, batch_id, bill_date, name, amount, date_init
		FROM bill_import_order WHERE uid=? AND batch_id=? ORDER BY id`, uid, batchID)
}

func loadBatchRows(path string) ([]RawRow, error) {
	raw, err := readFileBytes(path)
	if err != nil {
		return nil, err
	}
	text, err := decodeCSVBytes(raw)
	if err != nil {
		return nil, err
	}
	return parseCSVText(text)
}

func periodFromRows(rows []RawRow) (start, end *time.Time) {
	if len(rows) == 0 {
		return nil, nil
	}
	minT, maxT := rows[0].When, rows[0].When
	for _, r := range rows[1:] {
		if r.When.Before(minT) {
			minT = r.When
		}
		if r.When.After(maxT) {
			maxT = r.When
		}
	}
	return &minT, &maxT
}

func batchToMap(b *batchRow) map[string]interface{} {
	m := map[string]interface{}{
		"id":             b.ID,
		"filename":       b.Filename,
		"stored_path":    b.StoredPath,
		"row_count":      b.RowCount,
		"ready_count":    b.ReadyCount,
		"guess_count":    b.GuessCount,
		"confirm_count":  b.ConfirmCount,
		"unmapped_count": b.UnmappedCount,
		"skip_count":     b.SkipCount,
		"applied_count":  b.AppliedCount,
		"status":         b.Status,
	}
	if b.PeriodStart.Valid {
		m["period_start"] = b.PeriodStart.Time.Format("2006-01-02 15:04:05")
	}
	if b.PeriodEnd.Valid {
		m["period_end"] = b.PeriodEnd.Time.Format("2006-01-02 15:04:05")
	}
	if b.DateInit.Valid {
		m["date_init"] = b.DateInit.Time.Format("2006-01-02 15:04:05")
	}
	return m
}


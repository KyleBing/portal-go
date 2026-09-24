package wubi

import (
	"database/sql"
	"fmt"

	"github.com/KyleBing/portal-go/internal/db"
)

// MigratePlaintext 把五笔库改成 utf8mb4，并把已有的 base64 词库、\uXXXX emoji 词条写成原文。
func MigratePlaintext() error {
	wubi, err := db.Open(db.Wubi)
	if err != nil {
		return err
	}
	for _, table := range []string{"wubi_dict", "wubi_words", "wubi_category"} {
		if err := ensureUTF8MB4(wubi, table); err != nil {
			return err
		}
	}
	if err := rewriteColumn(wubi, dictTable, "id", "title"); err != nil {
		return err
	}
	if err := rewriteColumn(wubi, dictTable, "id", "content"); err != nil {
		return err
	}
	if err := rewriteColumn(wubi, wordTable, "id", "word"); err != nil {
		return err
	}
	return nil
}

func ensureUTF8MB4(sqlDB *sql.DB, table string) error {
	var charset, collation sql.NullString
	err := sqlDB.QueryRow(`
		SELECT CCSA.character_set_name, T.table_collation
		FROM information_schema.TABLES T
		JOIN information_schema.COLLATION_CHARACTER_SET_APPLICABILITY CCSA
		  ON CCSA.collation_name = T.table_collation
		WHERE T.table_schema = DATABASE() AND T.table_name = ?`, table).Scan(&charset, &collation)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s charset: %w", table, err)
	}
	if charset.String == "utf8mb4" {
		fmt.Printf("%-10s %s charset\n", "skip", table)
		return nil
	}
	// 旧表常是 COMPACT，utf8mb4 索引上限 767 字节，先改成 DYNAMIC 再转字符集
	if _, err := sqlDB.Exec(`ALTER TABLE ` + table + ` ROW_FORMAT=DYNAMIC`); err != nil {
		return fmt.Errorf("row format %s: %w", table, err)
	}
	if _, err := sqlDB.Exec(`ALTER TABLE ` + table + ` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci`); err != nil {
		return fmt.Errorf("alter %s: %w", table, err)
	}
	fmt.Printf("%-10s %s charset -> utf8mb4\n", "applied", table)
	return nil
}

func rewriteColumn(sqlDB *sql.DB, table, idCol, col string) error {
	rows, err := sqlDB.Query(`SELECT ` + idCol + `, ` + col + ` FROM ` + table + ` WHERE ` + col + ` IS NOT NULL AND ` + col + ` != ''`)
	if err != nil {
		return fmt.Errorf("scan %s.%s: %w", table, col, err)
	}
	defer rows.Close()

	type item struct {
		id   int64
		text string
	}
	var pending []item
	for rows.Next() {
		var id int64
		var text string
		if err := rows.Scan(&id, &text); err != nil {
			return err
		}
		plain := PlainText(text)
		if plain == text {
			continue
		}
		pending = append(pending, item{id: id, text: plain})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, it := range pending {
		if _, err := sqlDB.Exec(`UPDATE `+table+` SET `+col+` = ? WHERE `+idCol+` = ?`, it.text, it.id); err != nil {
			return fmt.Errorf("update %s.%s id=%d: %w", table, col, it.id, err)
		}
	}
	fmt.Printf("%-10s %s.%s %d rows\n", "rewritten", table, col, len(pending))
	return nil
}

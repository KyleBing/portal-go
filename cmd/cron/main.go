// Command cron is the Go port of portal/cron/updateUserInfo.ts. It refreshes the
// per-user statistics (diary / map / dict / qr / words counts) in the diary
// database. Intended to be scheduled hourly, e.g. via crontab:
//
//	17 * * * * /var/www/html/portal-go/bin/cron
package main

import (
	"log"

	"github.com/KyleBing/portal-go/internal/config"
	"github.com/KyleBing/portal-go/internal/db"
)

func main() {
	if _, err := config.Load(); err != nil {
		log.Fatalf("加载数据库配置失败: %v", err)
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		log.Fatalf("error: get users info: %v", err)
	}

	rows, err := diary.Query(`SELECT uid FROM users`)
	if err != nil {
		log.Fatalf("error: get users info: %v", err)
	}
	var uids []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err != nil {
			_ = rows.Close()
			log.Fatalf("error: get users info: %v", err)
		}
		uids = append(uids, uid)
	}
	_ = rows.Close()

	// Each statement mirrors a line from updateUserInfo.ts. wubi.* references
	// live in a separate database but are reachable via the fully-qualified name.
	statements := []string{
		`UPDATE users SET count_diary = (SELECT count(*) FROM diaries WHERE uid = ?) WHERE uid = ?`,
		`UPDATE users SET count_map_route = (SELECT count(*) FROM map_route WHERE uid = ?) WHERE uid = ?`,
		`UPDATE users SET count_map_pointer = (SELECT count(*) FROM map_pointer WHERE uid = ?) WHERE uid = ?`,
		`UPDATE users SET count_dict = (SELECT count(*) FROM wubi.wubi_dict WHERE uid = ?) WHERE uid = ?`,
		`UPDATE users SET count_qr = (SELECT count(*) FROM qrs WHERE uid = ?) WHERE uid = ?`,
		`UPDATE users SET count_words = (SELECT count(*) FROM wubi.wubi_words WHERE user_init = ? AND category_id != 1) WHERE uid = ?`,
	}

	failed := false
	for _, uid := range uids {
		for _, stmt := range statements {
			if _, err := diary.Exec(stmt, uid, uid); err != nil {
				failed = true
				log.Printf("error: user count update failed for uid=%d: %v", uid, err)
			}
		}
	}

	if failed {
		log.Println("error: user count diary|dict update")
	} else {
		log.Println("success: user's count diary|dict has updated")
	}
}

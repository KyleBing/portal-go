package util

import (
	"database/sql"

	"github.com/KyleBing/portal-go/internal/db"
)

func UpdateUserLastLoginTime(uid int64) {
	diary, err := db.Open(db.Diary)
	if err != nil {
		return
	}
	_, _ = diary.Exec(`UPDATE users SET last_visit_time = ? WHERE uid = ?`, NowString(), uid)
}

func NullStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func StrPtr(s string) *string {
	return &s
}

func StrOrEmpty(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

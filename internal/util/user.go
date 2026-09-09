package util

import (
	"database/sql"
	"sync"
	"time"

	"github.com/KyleBing/portal-go/internal/db"
)

// lastVisitMinInterval avoids writing last_visit_time on every authenticated request.
const lastVisitMinInterval = 5 * time.Minute

var (
	lastVisitMu   sync.Mutex
	lastVisitSeen = map[int64]time.Time{}
)

func UpdateUserLastLoginTime(uid int64) {
	if uid <= 0 {
		return
	}
	now := time.Now()
	lastVisitMu.Lock()
	if t, ok := lastVisitSeen[uid]; ok && now.Sub(t) < lastVisitMinInterval {
		lastVisitMu.Unlock()
		return
	}
	lastVisitSeen[uid] = now
	lastVisitMu.Unlock()

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

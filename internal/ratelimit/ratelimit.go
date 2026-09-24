package ratelimit

import (
	"sync"
	"time"
)

type entry struct {
	count   int
	resetAt time.Time
}

// Limiter 进程内简易限流（重启清空）。
type Limiter struct {
	mu   sync.Mutex
	data map[string]entry
}

func New() *Limiter {
	return &Limiter{data: make(map[string]entry)}
}

// Allow 在 window 内最多 max 次；超限返回 false。
func (l *Limiter) Allow(key string, max int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	e, ok := l.data[key]
	if !ok || now.After(e.resetAt) {
		l.data[key] = entry{count: 1, resetAt: now.Add(window)}
		return true
	}
	if e.count >= max {
		return false
	}
	e.count++
	l.data[key] = e
	return true
}

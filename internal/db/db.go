package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/KyleBing/portal-go/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

const (
	Diary         = "diary"
	Wubi          = "wubi"
	Starve        = "starve"
	StarveAdvance = "starve_advance"
)

var (
	mu    sync.RWMutex
	pools = map[string]*sql.DB{}
)

func dsn(dbName string) string {
	c := config.Get()
	loc := c.Timezone
	if loc == "" {
		loc = "Local"
	}
	// parseTime for DATETIME; multiStatements for init.sql
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=%s&multiStatements=true",
		c.User, c.Password, config.Host(), c.Port, dbName, loc)
}

func Open(dbName string) (*sql.DB, error) {
	mu.RLock()
	if p, ok := pools[dbName]; ok {
		mu.RUnlock()
		return p, nil
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()
	if p, ok := pools[dbName]; ok {
		return p, nil
	}
	p, err := sql.Open("mysql", dsn(dbName))
	if err != nil {
		return nil, err
	}
	p.SetMaxOpenConns(20)
	p.SetMaxIdleConns(5)
	p.SetConnMaxLifetime(time.Hour)
	if err := p.Ping(); err != nil {
		_ = p.Close()
		return nil, err
	}
	pools[dbName] = p
	return p, nil
}

func Must(dbName string) *sql.DB {
	p, err := Open(dbName)
	if err != nil {
		panic(err)
	}
	return p
}

func CloseAll() {
	mu.Lock()
	defer mu.Unlock()
	for k, p := range pools {
		_ = p.Close()
		delete(pools, k)
	}
}

// OpenWithoutDB connects without selecting a database (for CREATE DATABASE).
func OpenWithoutDB() (*sql.DB, error) {
	c := config.Get()
	loc := c.Timezone
	if loc == "" {
		loc = "Local"
	}
	s := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=true&loc=%s&multiStatements=true",
		c.User, c.Password, config.Host(), c.Port, loc)
	return sql.Open("mysql", s)
}

func ResetPools() {
	CloseAll()
}

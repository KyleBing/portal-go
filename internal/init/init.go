package initdb

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/setup"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup) {
	r.GET("/", handleInitHTML)
	r.GET("", handleInitHTML)
}

func handleInitHTML(c *gin.Context) {
	result, err := InitializeDatabase()
	if err != nil {
		response.Error(c, errData(err), err.Error())
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	if result.AlreadyInitialized {
		c.String(200, "%s", result.Message)
		return
	}
	c.String(200, "数据库初始化成功：<br>数据库名： %s<br>创建 6 张表：%s <br>已创建数据库锁定文件： %s",
		result.DBName, "users、user_group、diaries、diary_category、qrs、invitations", result.LockFileName)
}

func HandleInitJSON(c *gin.Context) {
	result, err := InitializeDatabase()
	if err != nil {
		response.Error(c, errData(err), err.Error())
		return
	}
	if result.AlreadyInitialized {
		response.Error(c, gin.H{
			"dbName":              result.DBName,
			"lockFileName":        result.LockFileName,
			"initializedByTables": setup.CoreTablesExist(),
			"initializedByLock":   setup.LockFileExists(),
			"allowSetupEnv":       setup.AllowSetupEnv(),
		}, result.Message)
		return
	}
	response.Success(c, gin.H{
		"dbName":       result.DBName,
		"tableNames":   []string{"users", "user_group", "diaries", "diary_category", "qrs", "invitations"},
		"lockFileName": result.LockFileName,
	}, result.Message)
}

type InitResult struct {
	AlreadyInitialized bool
	Message            string
	DBName             string
	LockFileName       string
	Stage              string
	Detail             string
}

func errData(err error) interface{} {
	if e, ok := err.(*InitError); ok {
		return gin.H{"stage": e.Stage, "step": e.Step, "detail": e.Detail}
	}
	return err.Error()
}

type InitError struct {
	Stage  string
	Step   string
	Detail string
}

func (e *InitError) Error() string {
	return fmt.Sprintf("%s失败：%s", e.Step, e.Detail)
}

func InitializeDatabase() (*InitResult, error) {
	if reason := setup.SetupBlockedReason(); reason != "" {
		return &InitResult{
			AlreadyInitialized: true,
			Message:            reason,
			DBName:             db.Diary,
			LockFileName:       setup.LockFileName,
		}, nil
	}

	conn, err := db.OpenWithoutDB()
	if err != nil {
		return nil, &InitError{Stage: "connect_mysql", Step: "连接数据库", Detail: err.Error()}
	}
	defer conn.Close()

	if _, err := conn.Exec("CREATE DATABASE IF NOT EXISTS diary"); err != nil {
		return nil, &InitError{Stage: "create_database", Step: "创建数据库", Detail: err.Error()}
	}

	sqlPath := findInitSQL()
	sqlBytes, err := os.ReadFile(sqlPath)
	if err != nil {
		return nil, &InitError{Stage: "create_tables", Step: "创建数据表", Detail: err.Error()}
	}

	diary, err := db.Open(db.Diary)
	if err != nil {
		return nil, &InitError{Stage: "connect_mysql", Step: "连接数据库", Detail: err.Error()}
	}
	if _, err := diary.Exec(string(sqlBytes)); err != nil {
		return nil, &InitError{Stage: "create_tables", Step: "创建数据表", Detail: err.Error()}
	}

	lockContent := "Database has been locked, file add in " + util.NowString()
	if err := os.WriteFile(setup.LockFilePath(), []byte(lockContent), 0o644); err != nil {
		return nil, &InitError{Stage: "create_lock_file", Step: "创建锁文件", Detail: err.Error()}
	}

	return &InitResult{
		AlreadyInitialized: false,
		Message:            "数据库初始化成功",
		DBName:             db.Diary,
		LockFileName:       setup.LockFileName,
	}, nil
}

func findInitSQL() string {
	candidates := []string{
		filepath.Join(setup.ProjectRoot(), "migrations", "init.sql"),
		"migrations/init.sql",
		filepath.Join("internal", "init", "init.sql"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return candidates[0]
}

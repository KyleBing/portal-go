package setup

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/KyleBing/portal-go/internal/config"
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const LockFileName = "DATABASE_LOCK"
const dbName = "diary"

// LockFilePath returns the path of the initialization lock file (relative to cwd /
// systemd WorkingDirectory). Kept for compatibility; initialization is primarily
// gated by whether core tables already exist.
func LockFilePath() string {
	return LockFileName
}

// LockFileExists reports whether DATABASE_LOCK is present.
func LockFileExists() bool {
	_, err := os.Stat(LockFilePath())
	return err == nil
}

// CoreTablesExist reports whether diary.users exists (source of truth for init).
func CoreTablesExist() bool {
	diary, err := db.Open(db.Diary)
	if err != nil {
		return false
	}
	var n int
	err = diary.QueryRow(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = ? AND table_name = 'users'`, db.Diary).Scan(&n)
	return err == nil && n > 0
}

// IsDatabaseInitialized is true when core tables exist or the legacy lock file is present.
func IsDatabaseInitialized() bool {
	return CoreTablesExist() || LockFileExists()
}

// IsInitialized is an alias for IsDatabaseInitialized.
func IsInitialized() bool { return IsDatabaseInitialized() }

// AllowSetupEnv reports whether ALLOW_SETUP permits the setup wizard.
// Explicit 0/false/no/off disables it (recommended in production).
// Unset or any other value leaves setup available until the DB is initialized.
func AllowSetupEnv() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("ALLOW_SETUP")))
	switch v {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

// CanMutateSetup is true only when env allows setup and the system is not initialized.
func CanMutateSetup() bool {
	return AllowSetupEnv() && !IsDatabaseInitialized()
}

// SetupBlockedReason explains why config/init mutations are refused (empty if allowed).
func SetupBlockedReason() string {
	if !AllowSetupEnv() {
		return "安装引导已禁用（ALLOW_SETUP=0）。如需重新安装，请由运维临时开启后再操作。"
	}
	if CoreTablesExist() {
		return "系统已初始化（检测到 diary.users 表），禁止重复执行初始化。"
	}
	if LockFileExists() {
		return fmt.Sprintf("系统已初始化（存在 %s）。如确认库表已清空，删除该文件后重试。", LockFileName)
	}
	return ""
}

// InitFn is the JSON database-initialization handler, wired by the server so
// the /setup/init route can share the /init implementation.
var InitFn gin.HandlerFunc

// ProjectRoot returns the directory that contains go.mod, falling back to the
// current working directory. It is used to locate migrations, uploads and the
// bundled manager frontend.
func ProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	d := dir
	for {
		if _, statErr := os.Stat(filepath.Join(d, "go.mod")); statErr == nil {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}
	return dir
}

func GetDatabaseConfig() config.DatabaseConfig {
	return config.Get()
}

func hasRegisteredUsers() bool {
	if !CoreTablesExist() {
		return false
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		return false
	}
	var count int64
	if err := diary.QueryRow(`select count(*) as userCount from users`).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

// GetSetupStatus mirrors setupService.getSetupStatus.
func GetSetupStatus() gin.H {
	initialized := IsDatabaseInitialized()
	var cfg interface{}
	if !initialized && AllowSetupEnv() {
		cfg = gin.H{"databaseConfig": GetDatabaseConfig()}
	}
	return gin.H{
		"isInitialized":       initialized,
		"initializedByTables": CoreTablesExist(),
		"initializedByLock":   LockFileExists(),
		"allowSetup":          CanMutateSetup(),
		"allowSetupEnv":       AllowSetupEnv(),
		"hasRegisteredUsers":  hasRegisteredUsers(),
		"lockFileName":        LockFileName,
		"configFiles": []string{
			"config/configDatabase.json",
			"dist/config/configDatabase.json",
		},
		"config": cfg,
		"restartTips": []string{
			"保存后会同步写入配置文件，并立即更新当前服务进程内存中的配置。",
			"项目配置、通用邀请码和七牛后台密钥请在初始化完成后，到系统配置页中维护。",
			"为了确保后续重启后的运行结果与当前一致，建议完成向导后重启 portal 服务。",
			"生产环境请设置 ALLOW_SETUP=0，初始化判定以 diary.users 表为准。",
		},
	}
}

// SaveSetupConfig persists the database config; only allowed before init.
func SaveSetupConfig(body map[string]interface{}) (gin.H, error) {
	if reason := SetupBlockedReason(); reason != "" {
		return nil, errors.New(reason)
	}
	raw, _ := body["databaseConfig"].(map[string]interface{})
	if raw == nil {
		raw = map[string]interface{}{}
	}
	current := config.Get()
	newCfg := config.DatabaseConfig{
		Host:               strings.TrimSpace(toStr(raw["host"])),
		User:               strings.TrimSpace(toStr(raw["user"])),
		Password:           toStr(raw["password"]),
		Port:               toInt(raw["port"], 3306),
		MultipleStatements: current.MultipleStatements,
		Timezone:           strings.TrimSpace(toStr(raw["timezone"])),
	}
	if v, ok := raw["multipleStatements"]; ok {
		newCfg.MultipleStatements = toBool(v)
	}
	if newCfg.Host == "" {
		return nil, errors.New("数据库主机不能为空")
	}
	if newCfg.User == "" {
		return nil, errors.New("数据库用户名不能为空")
	}
	if newCfg.Port <= 0 || newCfg.Port > 65535 {
		return nil, errors.New("数据库端口不正确")
	}
	if err := config.Save(newCfg); err != nil {
		return nil, err
	}
	db.ResetPools()
	return gin.H{"databaseConfig": config.Get()}, nil
}

// InitResult is the result of InitializeDatabase.
type InitResult struct {
	AlreadyInitialized bool
	Message            string
	Data               gin.H
}

func readInitSQL() (string, error) {
	candidates := []string{
		"migrations/init.sql",
		filepath.Join("..", "migrations", "init.sql"),
		filepath.Join("..", "..", "migrations", "init.sql"),
	}
	for _, p := range candidates {
		if data, err := os.ReadFile(p); err == nil {
			return string(data), nil
		}
	}
	return "", errors.New("找不到 migrations/init.sql")
}

// InitializeDatabase mirrors initService.initializeDatabase.
func InitializeDatabase() (InitResult, error) {
	if reason := SetupBlockedReason(); reason != "" {
		return InitResult{
			AlreadyInitialized: true,
			Message:            reason,
			Data: gin.H{
				"dbName":              dbName,
				"lockFileName":        LockFileName,
				"initializedByTables": CoreTablesExist(),
				"initializedByLock":   LockFileExists(),
				"allowSetupEnv":       AllowSetupEnv(),
			},
		}, nil
	}

	rootDB, err := db.OpenWithoutDB()
	if err != nil {
		return InitResult{}, fmt.Errorf("连接数据库失败：%v", err)
	}
	defer rootDB.Close()
	if _, err := rootDB.Exec("CREATE DATABASE IF NOT EXISTS " + dbName); err != nil {
		return InitResult{}, fmt.Errorf("创建数据库失败：%v", err)
	}

	sqlText, err := readInitSQL()
	if err != nil {
		return InitResult{}, fmt.Errorf("创建数据表失败：%v", err)
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		return InitResult{}, fmt.Errorf("连接数据库失败：%v", err)
	}
	if _, err := diary.Exec(sqlText); err != nil {
		return InitResult{}, fmt.Errorf("创建数据表失败：%v", err)
	}

	if err := os.WriteFile(LockFilePath(), []byte("Database has been locked, file add in "+util.NowString()), 0o644); err != nil {
		return InitResult{}, fmt.Errorf("创建锁文件失败：%v", err)
	}

	return InitResult{
		AlreadyInitialized: false,
		Message:            "数据库初始化成功",
		Data: gin.H{
			"dbName":       dbName,
			"tableNames":   []string{"users", "user_group", "diaries", "diary_category", "qrs", "invitations"},
			"lockFileName": LockFileName,
		},
	}, nil
}

// FormatInitResultHtml mirrors initService.formatInitResultHtml.
func FormatInitResultHtml(r InitResult) string {
	if r.AlreadyInitialized {
		return r.Message
	}
	names, _ := r.Data["tableNames"].([]string)
	return "数据库初始化成功：<br>" +
		fmt.Sprintf("数据库名： %v<br>", r.Data["dbName"]) +
		fmt.Sprintf("创建 6 张表：%s <br>", strings.Join(names, "、")) +
		fmt.Sprintf("已创建数据库锁定文件： %v", r.Data["lockFileName"])
}

// Register mounts the setup routes.
func Register(r *gin.RouterGroup) {
	g := r

	g.GET("/status", func(c *gin.Context) {
		response.Success(c, GetSetupStatus(), "请求成功")
	})

	g.POST("/config", func(c *gin.Context) {
		var body map[string]interface{}
		_ = c.ShouldBindJSON(&body)
		if body == nil {
			body = map[string]interface{}{}
		}
		data, err := SaveSetupConfig(body)
		if err != nil {
			response.Error(c, err.Error(), err.Error())
			return
		}
		response.Success(c, data, "配置已保存，当前服务已同步新配置，建议随后重启服务")
	})

	g.POST("/init", func(c *gin.Context) {
		if reason := SetupBlockedReason(); reason != "" {
			response.Error(c, gin.H{
				"dbName":              dbName,
				"lockFileName":        LockFileName,
				"initializedByTables": CoreTablesExist(),
				"initializedByLock":   LockFileExists(),
				"allowSetupEnv":       AllowSetupEnv(),
			}, reason)
			return
		}
		if InitFn != nil {
			InitFn(c)
			return
		}
		res, err := InitializeDatabase()
		if err != nil {
			response.Error(c, err.Error(), err.Error())
			return
		}
		if res.AlreadyInitialized {
			response.Error(c, res.Data, res.Message)
			return
		}
		response.Success(c, res.Data, res.Message)
	})
}

func toStr(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func toInt(v interface{}, def int) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case string:
		if t == "" {
			return def
		}
		var n int
		_, err := fmt.Sscanf(t, "%d", &n)
		if err != nil {
			return def
		}
		return n
	}
	return def
}

func toBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case string:
		return t == "1" || t == "true"
	}
	return false
}

var _ = sql.ErrNoRows

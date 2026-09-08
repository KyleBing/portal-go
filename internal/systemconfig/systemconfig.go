package systemconfig

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/setup"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const tableName = "system_config"
const usersTable = "users"

// SystemConfig holds the full config (with secrets).
type SystemConfig struct {
	IsShowDemoAccount    bool   `json:"is_show_demo_account"`
	DemoAccount          string `json:"demo_account"`
	DemoAccountPassword  string `json:"demo_account_password"`
	InvitationCode       string `json:"invitation_code"`
	QiniuImgBaseURL      string `json:"qiniu_img_base_url"`
	QiniuBucketName      string `json:"qiniu_bucket_name"`
	QiniuStyleSuffix     string `json:"qiniu_style_suffix"`
	QiniuAccessKey       string `json:"qiniu_access_key"`
	QiniuSecretKey       string `json:"qiniu_secret_key"`
	HefengWeatherAPIKey  string `json:"hefeng_weather_api_key"`
	HefengWeatherAPIHost string `json:"hefeng_weather_api_host"`
	RegisterTip          string `json:"register_tip"`
}

var defaultConfig = SystemConfig{
	IsShowDemoAccount:    false,
	DemoAccount:          "test@163.com",
	DemoAccountPassword:  "test",
	InvitationCode:       "",
	QiniuImgBaseURL:      "http://diary-container.kylebing.cn/",
	QiniuBucketName:      "diary-container",
	QiniuStyleSuffix:     "thumbnail_600px",
	QiniuAccessKey:       "",
	QiniuSecretKey:       "",
	HefengWeatherAPIKey:  "c5894aea6ce2495ca0f78a2963c04d57",
	HefengWeatherAPIHost: "pd3fbqjryn.re.qweatherapi.com",
	RegisterTip:          "<p>长期未使用的用户将定期进行清理，大概一年清一次。</p><p>项目已开源</p>",
}

// PublicConfig strips secrets.
func (c SystemConfig) Public() gin.H {
	return gin.H{
		"is_show_demo_account":    c.IsShowDemoAccount,
		"demo_account":            c.DemoAccount,
		"demo_account_password":   c.DemoAccountPassword,
		"qiniu_img_base_url":      c.QiniuImgBaseURL,
		"qiniu_bucket_name":       c.QiniuBucketName,
		"qiniu_style_suffix":      c.QiniuStyleSuffix,
		"hefeng_weather_api_key":  c.HefengWeatherAPIKey,
		"hefeng_weather_api_host": c.HefengWeatherAPIHost,
		"register_tip":            c.RegisterTip,
	}
}

func trimAll(c SystemConfig) SystemConfig {
	c.DemoAccount = strings.TrimSpace(c.DemoAccount)
	c.InvitationCode = strings.TrimSpace(c.InvitationCode)
	c.QiniuImgBaseURL = strings.TrimSpace(c.QiniuImgBaseURL)
	c.QiniuBucketName = strings.TrimSpace(c.QiniuBucketName)
	c.QiniuStyleSuffix = strings.TrimSpace(c.QiniuStyleSuffix)
	c.QiniuAccessKey = strings.TrimSpace(c.QiniuAccessKey)
	c.QiniuSecretKey = strings.TrimSpace(c.QiniuSecretKey)
	c.HefengWeatherAPIKey = strings.TrimSpace(c.HefengWeatherAPIKey)
	c.HefengWeatherAPIHost = strings.TrimSpace(c.HefengWeatherAPIHost)
	c.RegisterTip = strings.TrimSpace(c.RegisterTip)
	return c
}

func ensureTable(diary *sql.DB) error {
	_, err := diary.Exec(`
        CREATE TABLE IF NOT EXISTS ` + tableName + ` (
            id tinyint(1) NOT NULL DEFAULT 1 COMMENT '固定主键，仅保留一行',
            is_show_demo_account tinyint(1) NOT NULL DEFAULT 0,
            demo_account varchar(255) NOT NULL DEFAULT '',
            demo_account_password varchar(255) NOT NULL DEFAULT '',
            invitation_code varchar(255) NOT NULL DEFAULT '',
            qiniu_img_base_url varchar(255) NOT NULL DEFAULT '',
            qiniu_bucket_name varchar(255) NOT NULL DEFAULT '',
            qiniu_style_suffix varchar(255) NOT NULL DEFAULT '',
            qiniu_access_key varchar(255) NOT NULL DEFAULT '',
            qiniu_secret_key varchar(255) NOT NULL DEFAULT '',
            hefeng_weather_api_key varchar(255) NOT NULL DEFAULT '',
            hefeng_weather_api_host varchar(255) NOT NULL DEFAULT '',
            register_tip text NULL,
            date_modify datetime NULL DEFAULT NULL,
            PRIMARY KEY (id)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci
    `)
	if err != nil {
		return err
	}
	// insert default row if missing
	_, err = diary.Exec(`INSERT IGNORE INTO `+tableName+` (
            id, is_show_demo_account, demo_account, demo_account_password, invitation_code,
            qiniu_img_base_url, qiniu_bucket_name, qiniu_style_suffix, qiniu_access_key, qiniu_secret_key,
            hefeng_weather_api_key, hefeng_weather_api_host, register_tip, date_modify
        ) VALUES (1, ?,?,?,?,?,?,?,?,?,?,?,?,NULL)`,
		boolToInt(defaultConfig.IsShowDemoAccount), defaultConfig.DemoAccount, defaultConfig.DemoAccountPassword,
		defaultConfig.InvitationCode, defaultConfig.QiniuImgBaseURL, defaultConfig.QiniuBucketName,
		defaultConfig.QiniuStyleSuffix, defaultConfig.QiniuAccessKey, defaultConfig.QiniuSecretKey,
		defaultConfig.HefengWeatherAPIKey, defaultConfig.HefengWeatherAPIHost, defaultConfig.RegisterTip,
	)
	if err != nil {
		return err
	}
	migrateLegacyProjectConfig(diary)
	return nil
}

func migrateLegacyProjectConfig(diary *sql.DB) {
	paths := []string{"config/configProject.json", "../config/configProject.json"}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var raw struct {
			InvitationCode string `json:"invitation_code"`
			QiniuAccessKey string `json:"qiniu_access_key"`
			QiniuSecretKey string `json:"qiniu_secret_key"`
		}
		if json.Unmarshal(data, &raw) != nil {
			continue
		}
		if raw.InvitationCode == "" && raw.QiniuAccessKey == "" && raw.QiniuSecretKey == "" {
			return
		}
		_, _ = diary.Exec(`UPDATE `+tableName+` SET
            invitation_code = CASE WHEN invitation_code = '' THEN ? ELSE invitation_code END,
            qiniu_access_key = CASE WHEN qiniu_access_key = '' THEN ? ELSE qiniu_access_key END,
            qiniu_secret_key = CASE WHEN qiniu_secret_key = '' THEN ? ELSE qiniu_secret_key END
            WHERE id = 1`, raw.InvitationCode, raw.QiniuAccessKey, raw.QiniuSecretKey)
		return
	}
}

func scanConfig(diary *sql.DB) (SystemConfig, error) {
	cfg := defaultConfig
	var isShow int
	var demoAccount, demoPwd, invite, imgBase, bucket, style, ak, sk, wkey, whost sql.NullString
	var registerTip sql.NullString
	row := diary.QueryRow(`SELECT is_show_demo_account, demo_account, demo_account_password, invitation_code,
        qiniu_img_base_url, qiniu_bucket_name, qiniu_style_suffix, qiniu_access_key, qiniu_secret_key,
        hefeng_weather_api_key, hefeng_weather_api_host, register_tip FROM ` + tableName + ` WHERE id = 1 LIMIT 1`)
	err := row.Scan(&isShow, &demoAccount, &demoPwd, &invite, &imgBase, &bucket, &style, &ak, &sk, &wkey, &whost, &registerTip)
	if err == sql.ErrNoRows {
		return defaultConfig, nil
	}
	if err != nil {
		return cfg, err
	}
	cfg = SystemConfig{
		IsShowDemoAccount:    isShow != 0,
		DemoAccount:          demoAccount.String,
		DemoAccountPassword:  demoPwd.String,
		InvitationCode:       invite.String,
		QiniuImgBaseURL:      imgBase.String,
		QiniuBucketName:      bucket.String,
		QiniuStyleSuffix:     style.String,
		QiniuAccessKey:       ak.String,
		QiniuSecretKey:       sk.String,
		HefengWeatherAPIKey:  wkey.String,
		HefengWeatherAPIHost: whost.String,
		RegisterTip:          registerTip.String,
	}
	return trimAll(cfg), nil
}

// GetAdminSystemConfig returns the full config (with secrets).
func GetAdminSystemConfig() (SystemConfig, error) {
	if !setup.IsDatabaseInitialized() {
		return trimAll(defaultConfig), nil
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		return defaultConfig, err
	}
	if err := ensureTable(diary); err != nil {
		return defaultConfig, err
	}
	return scanConfig(diary)
}

// GetSystemConfig returns the public config.
func GetSystemConfig() (gin.H, error) {
	if !setup.IsDatabaseInitialized() {
		return trimAll(defaultConfig).Public(), nil
	}
	cfg, err := GetAdminSystemConfig()
	if err != nil {
		return nil, err
	}
	return cfg.Public(), nil
}

// GetAdmin is an alias for GetAdminSystemConfig.
func GetAdmin() (SystemConfig, error) { return GetAdminSystemConfig() }

// IsDemoEmail is an alias for IsConfiguredDemoAccountEmail.
func IsDemoEmail(email string) bool { return IsConfiguredDemoAccountEmail(email) }

// IsConfiguredDemoAccountEmail reports whether the email matches the demo account.
func IsConfiguredDemoAccountEmail(email string) bool {
	cfg, err := GetSystemConfig()
	if err != nil {
		return false
	}
	demo, _ := cfg["demo_account"].(string)
	demo = strings.TrimSpace(demo)
	return demo != "" && email == demo
}

// SaveSystemConfig merges and persists config.
func SaveSystemConfig(payload map[string]interface{}) (SystemConfig, error) {
	if !setup.IsDatabaseInitialized() {
		return SystemConfig{}, errors.New("系统尚未初始化，暂时不能保存系统配置")
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		return SystemConfig{}, err
	}
	if err := ensureTable(diary); err != nil {
		return SystemConfig{}, err
	}
	current, err := scanConfig(diary)
	if err != nil {
		return SystemConfig{}, err
	}
	merged := mergePayload(current, payload)
	merged = trimAll(merged)
	if merged.IsShowDemoAccount {
		if merged.DemoAccount == "" {
			return SystemConfig{}, errors.New("启用演示账号时，演示账号邮箱不能为空")
		}
		if len(merged.DemoAccount) > 50 {
			return SystemConfig{}, errors.New("演示账号邮箱长度不能超过 50 个字符")
		}
		if merged.DemoAccountPassword == "" {
			return SystemConfig{}, errors.New("启用演示账号时，演示账号密码不能为空")
		}
	}
	_, err = diary.Exec(`INSERT INTO `+tableName+` (
            id, is_show_demo_account, demo_account, demo_account_password, invitation_code,
            qiniu_img_base_url, qiniu_bucket_name, qiniu_style_suffix, qiniu_access_key, qiniu_secret_key,
            hefeng_weather_api_key, hefeng_weather_api_host, register_tip, date_modify
        ) VALUES (1,?,?,?,?,?,?,?,?,?,?,?,?,?)
        ON DUPLICATE KEY UPDATE
            is_show_demo_account=VALUES(is_show_demo_account),
            demo_account=VALUES(demo_account),
            demo_account_password=VALUES(demo_account_password),
            invitation_code=VALUES(invitation_code),
            qiniu_img_base_url=VALUES(qiniu_img_base_url),
            qiniu_bucket_name=VALUES(qiniu_bucket_name),
            qiniu_style_suffix=VALUES(qiniu_style_suffix),
            qiniu_access_key=VALUES(qiniu_access_key),
            qiniu_secret_key=VALUES(qiniu_secret_key),
            hefeng_weather_api_key=VALUES(hefeng_weather_api_key),
            hefeng_weather_api_host=VALUES(hefeng_weather_api_host),
            register_tip=VALUES(register_tip),
            date_modify=VALUES(date_modify)`,
		boolToInt(merged.IsShowDemoAccount), merged.DemoAccount, merged.DemoAccountPassword, merged.InvitationCode,
		merged.QiniuImgBaseURL, merged.QiniuBucketName, merged.QiniuStyleSuffix, merged.QiniuAccessKey, merged.QiniuSecretKey,
		merged.HefengWeatherAPIKey, merged.HefengWeatherAPIHost, merged.RegisterTip, util.NowString())
	if err != nil {
		return SystemConfig{}, err
	}
	if merged.IsShowDemoAccount {
		syncDemoAccountUser(diary, merged)
	}
	return scanConfig(diary)
}

func mergePayload(cur SystemConfig, p map[string]interface{}) SystemConfig {
	getStr := func(k, def string) string {
		if v, ok := p[k]; ok && v != nil {
			if s, ok := v.(string); ok {
				return s
			}
			return fmt.Sprint(v)
		}
		return def
	}
	getBool := func(k string, def bool) bool {
		if v, ok := p[k]; ok {
			switch t := v.(type) {
			case bool:
				return t
			case float64:
				return t == 1
			case string:
				return t == "1" || t == "true"
			}
		}
		return def
	}
	return SystemConfig{
		IsShowDemoAccount:    getBool("is_show_demo_account", cur.IsShowDemoAccount),
		DemoAccount:          getStr("demo_account", cur.DemoAccount),
		DemoAccountPassword:  getStr("demo_account_password", cur.DemoAccountPassword),
		InvitationCode:       getStr("invitation_code", cur.InvitationCode),
		QiniuImgBaseURL:      getStr("qiniu_img_base_url", cur.QiniuImgBaseURL),
		QiniuBucketName:      getStr("qiniu_bucket_name", cur.QiniuBucketName),
		QiniuStyleSuffix:     getStr("qiniu_style_suffix", cur.QiniuStyleSuffix),
		QiniuAccessKey:       getStr("qiniu_access_key", cur.QiniuAccessKey),
		QiniuSecretKey:       getStr("qiniu_secret_key", cur.QiniuSecretKey),
		HefengWeatherAPIKey:  getStr("hefeng_weather_api_key", cur.HefengWeatherAPIKey),
		HefengWeatherAPIHost: getStr("hefeng_weather_api_host", cur.HefengWeatherAPIHost),
		RegisterTip:          getStr("register_tip", cur.RegisterTip),
	}
}

var nonUsernameRe = regexp.MustCompile(`[^a-z0-9_]`)

func syncDemoAccountUser(diary *sql.DB, cfg SystemConfig) {
	email := strings.TrimSpace(cfg.DemoAccount)
	plain := cfg.DemoAccountPassword

	var uid int64
	var pwd string
	var groupID int
	err := diary.QueryRow(`SELECT uid, password, group_id FROM `+usersTable+` WHERE email = ? LIMIT 1`, email).Scan(&uid, &pwd, &groupID)
	if err == sql.ErrNoRows {
		hash, herr := bcrypt.GenerateFromPassword([]byte(plain), 10)
		if herr != nil {
			return
		}
		now := util.NowString()
		nickname := deriveDemoNickname(email)
		username := pickUniqueDemoUsername(diary, deriveDemoUsernameBase(email))
		_, _ = diary.Exec(`INSERT INTO `+usersTable+`(email, nickname, username, password, register_time, last_visit_time, comment, wx, phone, homepage, gaode, group_id)
            VALUES (?,?,?,?,?,?,'','','','','',2)`, email, nickname, username, string(hash), now, now)
		return
	}
	if err != nil {
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(pwd), []byte(plain)) == nil {
		return
	}
	hash, herr := bcrypt.GenerateFromPassword([]byte(plain), 10)
	if herr != nil {
		return
	}
	_, _ = diary.Exec(`UPDATE `+usersTable+` SET password = ? WHERE uid = ? LIMIT 1`, string(hash), uid)
}

func deriveDemoNickname(email string) string {
	local := email
	if i := strings.Index(email, "@"); i >= 0 {
		local = email[:i]
	}
	if local == "" {
		local = "演示"
	}
	r := []rune(local)
	if len(r) > 20 {
		r = r[:20]
	}
	return string(r)
}

func deriveDemoUsernameBase(email string) string {
	local := email
	if i := strings.Index(email, "@"); i >= 0 {
		local = email[:i]
	}
	if local == "" {
		local = "demo"
	}
	sanitized := nonUsernameRe.ReplaceAllString(strings.ToLower(local), "")
	if sanitized == "" {
		sanitized = "demo"
	}
	if len(sanitized) > 20 {
		sanitized = sanitized[:20]
	}
	return sanitized
}

func pickUniqueDemoUsername(diary *sql.DB, base string) string {
	if len(base) > 20 {
		base = base[:20]
	}
	for i := 0; i < 10000; i++ {
		suffix := ""
		if i > 0 {
			suffix = fmt.Sprintf("%d", i)
		}
		maxBase := 20 - len(suffix)
		if maxBase < 1 {
			maxBase = 1
		}
		b := base
		if len(b) > maxBase {
			b = b[:maxBase]
		}
		candidate := b + suffix
		if len(candidate) > 20 {
			candidate = candidate[:20]
		}
		var x int64
		err := diary.QueryRow(`SELECT uid FROM `+usersTable+` WHERE username = ? LIMIT 1`, candidate).Scan(&x)
		if err == sql.ErrNoRows {
			return candidate
		}
	}
	return base
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Register mounts system-config routes.
func Register(r *gin.RouterGroup) {
	g := r

	g.GET("", func(c *gin.Context) {
		data, err := GetSystemConfig()
		if err != nil {
			response.Error(c, err.Error(), "读取系统配置失败")
			return
		}
		response.Success(c, data, "请求成功")
	})

	g.GET("/admin", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		data, err := GetAdminSystemConfig()
		if err != nil {
			response.Error(c, err.Error(), "读取系统配置失败")
			return
		}
		response.Success(c, data, "请求成功")
	})

	g.PUT("", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		var body map[string]interface{}
		_ = c.ShouldBindJSON(&body)
		if body == nil {
			body = map[string]interface{}{}
		}
		data, err := SaveSystemConfig(body)
		if err != nil {
			response.Error(c, err.Error(), err.Error())
			return
		}
		response.Success(c, data, "系统配置已保存")
	})
}

func requireAdmin(c *gin.Context) bool {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", errMsg)
		return false
	}
	if !user.IsAdmin() {
		response.Error(c, "", "仅管理员可操作系统配置")
		return false
	}
	return true
}

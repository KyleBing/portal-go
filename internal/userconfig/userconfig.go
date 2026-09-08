package userconfig

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/models"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const tableName = "user_config"

// UserConfig mirrors the Node portal user_config row.
type UserConfig struct {
	UID                   int64                  `json:"uid"`
	Theme                 string                 `json:"theme"`
	DefaultDiaryCategory  string                 `json:"default_diary_category"`
	EditorMode            string                 `json:"editor_mode"`
	ConfigJSON            map[string]interface{} `json:"config_json"`
	DateModify            *string                `json:"date_modify"`
}

// Register mounts /user-config routes.
func Register(r *gin.RouterGroup) {
	r.GET("/", handleGet)
	r.PUT("/", handleSave)
}

func handleGet(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, errMsg, errMsg)
		return
	}
	cfg, err := getUserConfig(user)
	if err != nil {
		response.Error(c, err.Error(), "读取用户配置失败")
		return
	}
	response.Success(c, cfg, "请求成功")
}

func handleSave(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, errMsg, errMsg)
		return
	}
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil && err.Error() != "EOF" {
		response.Error(c, err.Error(), "保存用户配置失败")
		return
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	cfg, err := saveUserConfig(user, payload)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	response.Success(c, cfg, "用户配置已保存")
}

func ensureTable(diary *sql.DB) error {
	_, err := diary.Exec(`
		CREATE TABLE IF NOT EXISTS ` + tableName + ` (
			uid int(11) NOT NULL COMMENT '用户 ID',
			theme varchar(20) NOT NULL DEFAULT '' COMMENT '主题',
			default_diary_category varchar(50) NOT NULL DEFAULT '' COMMENT '默认日记分类',
			editor_mode varchar(20) NOT NULL DEFAULT '' COMMENT '编辑器模式',
			config_json json DEFAULT NULL COMMENT '扩展配置',
			date_modify datetime DEFAULT NULL COMMENT '最后修改时间',
			PRIMARY KEY (uid) USING BTREE,
			CONSTRAINT user_config_uid FOREIGN KEY (uid) REFERENCES users (uid) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci
	`)
	return err
}

func parseConfigJSON(value interface{}) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	switch v := value.(type) {
	case map[string]interface{}:
		return v
	case []byte:
		var out map[string]interface{}
		if err := json.Unmarshal(v, &out); err != nil || out == nil {
			return map[string]interface{}{}
		}
		return out
	case string:
		if strings.TrimSpace(v) == "" {
			return map[string]interface{}{}
		}
		var out map[string]interface{}
		if err := json.Unmarshal([]byte(v), &out); err != nil || out == nil {
			return map[string]interface{}{}
		}
		return out
	default:
		return map[string]interface{}{}
	}
}

func normalize(uid int64, raw UserConfig) UserConfig {
	cfg := UserConfig{
		UID:                  uid,
		Theme:                strings.TrimSpace(raw.Theme),
		DefaultDiaryCategory: strings.TrimSpace(raw.DefaultDiaryCategory),
		EditorMode:           strings.TrimSpace(raw.EditorMode),
		ConfigJSON:           parseConfigJSON(raw.ConfigJSON),
		DateModify:           raw.DateModify,
	}
	if cfg.ConfigJSON == nil {
		cfg.ConfigJSON = map[string]interface{}{}
	}
	return cfg
}

func asTrimmedString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func validate(cfg UserConfig) error {
	if len(cfg.Theme) > 20 {
		return fmt.Errorf("主题长度不能超过 20 个字符")
	}
	if len(cfg.DefaultDiaryCategory) > 50 {
		return fmt.Errorf("默认日记分类长度不能超过 50 个字符")
	}
	if len(cfg.EditorMode) > 20 {
		return fmt.Errorf("编辑器模式长度不能超过 20 个字符")
	}
	return nil
}

func getUserConfig(user *models.User) (UserConfig, error) {
	diary, err := db.Open(db.Diary)
	if err != nil {
		return UserConfig{}, err
	}
	if err := ensureTable(diary); err != nil {
		return UserConfig{}, err
	}

	var theme, category, editor sql.NullString
	var configRaw []byte
	var dateModify sql.NullString
	err = diary.QueryRow(
		`SELECT theme, default_diary_category, editor_mode, config_json, date_modify FROM `+tableName+` WHERE uid = ? LIMIT 1`,
		user.UID,
	).Scan(&theme, &category, &editor, &configRaw, &dateModify)
	if err == sql.ErrNoRows {
		return normalize(user.UID, UserConfig{}), nil
	}
	if err != nil {
		return UserConfig{}, err
	}

	raw := UserConfig{
		Theme:                theme.String,
		DefaultDiaryCategory: category.String,
		EditorMode:           editor.String,
		ConfigJSON:           parseConfigJSON(configRaw),
	}
	if dateModify.Valid {
		s := dateModify.String
		raw.DateModify = &s
	}
	return normalize(user.UID, raw), nil
}

func saveUserConfig(user *models.User, payload map[string]interface{}) (UserConfig, error) {
	current, err := getUserConfig(user)
	if err != nil {
		return UserConfig{}, err
	}

	if _, ok := payload["theme"]; ok {
		current.Theme = asTrimmedString(payload["theme"])
	}
	if _, ok := payload["default_diary_category"]; ok {
		current.DefaultDiaryCategory = asTrimmedString(payload["default_diary_category"])
	}
	if _, ok := payload["editor_mode"]; ok {
		current.EditorMode = asTrimmedString(payload["editor_mode"])
	}
	if _, ok := payload["config_json"]; ok {
		current.ConfigJSON = parseConfigJSON(payload["config_json"])
	}
	current = normalize(user.UID, current)
	if err := validate(current); err != nil {
		return UserConfig{}, err
	}

	diary, err := db.Open(db.Diary)
	if err != nil {
		return UserConfig{}, err
	}
	if err := ensureTable(diary); err != nil {
		return UserConfig{}, err
	}

	configBytes, err := json.Marshal(current.ConfigJSON)
	if err != nil {
		return UserConfig{}, err
	}
	now := util.NowString()
	_, err = diary.Exec(`
		INSERT INTO `+tableName+` (
			uid, theme, default_diary_category, editor_mode, config_json, date_modify
		) VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			theme = VALUES(theme),
			default_diary_category = VALUES(default_diary_category),
			editor_mode = VALUES(editor_mode),
			config_json = VALUES(config_json),
			date_modify = VALUES(date_modify)
	`, current.UID, current.Theme, current.DefaultDiaryCategory, current.EditorMode, string(configBytes), now)
	if err != nil {
		return UserConfig{}, err
	}
	return getUserConfig(user)
}

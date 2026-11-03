package payqr

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/setup"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const tableName = "registration_pay_qr_items"
const relPathPrefix = "registration-pay-qr/file"

var payQrContentTypeExt = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// Item 一档金额对应的两张收款码及备注。
type Item struct {
	ID         int64  `json:"id"`
	AmountYuan int    `json:"amount_yuan"`
	AlipayURL  string `json:"alipay_url"`
	AlipayNote string `json:"alipay_note"`
	WechatURL  string `json:"wechat_url"`
	WechatNote string `json:"wechat_note"`
	DateCreate string `json:"date_create"`
	DateModify string `json:"date_modify"`
}

// Register 挂载收款码列表路由。
func Register(r *gin.RouterGroup) {
	r.GET("", handleList)
	r.POST("", handleCreate)
	r.PUT("", handleUpdate)
	r.DELETE("", handleDelete)
	r.GET("/file/:id/:platform", handleFileGet)
	r.POST("/:id/upload/:platform", handleUpload)
}

func diaryDB() (*sql.DB, error) { return db.Open(db.Diary) }

func payQrDir() string {
	return filepath.Join(setup.ProjectRoot(), "data", "pay-qr")
}

func ensureTable(diary *sql.DB) error {
	_, err := diary.Exec(`
        CREATE TABLE IF NOT EXISTS ` + tableName + ` (
            id bigint unsigned NOT NULL AUTO_INCREMENT,
            amount_yuan int NOT NULL COMMENT '应付金额（元），0 表示自定义金额',
            alipay_url varchar(1024) NOT NULL DEFAULT '',
            alipay_note varchar(255) NOT NULL DEFAULT '',
            wechat_url varchar(1024) NOT NULL DEFAULT '',
            wechat_note varchar(255) NOT NULL DEFAULT '',
            date_create datetime NOT NULL,
            date_modify datetime DEFAULT NULL,
            PRIMARY KEY (id),
            UNIQUE KEY uk_reg_pay_qr_amount (amount_yuan)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
    `)
	return err
}

func normalizePlatform(platform string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "alipay":
		return "alipay", true
	case "wechat", "wepay":
		return "wechat", true
	default:
		return "", false
	}
}

func IsValidAmount(n int) bool {
	return n >= 0 && n <= 100000
}

// ListItems 按金额排序，自定义金额（0）放最后。
func ListItems() ([]Item, error) {
	if !setup.IsDatabaseInitialized() {
		return []Item{}, nil
	}
	diary, err := diaryDB()
	if err != nil {
		return nil, err
	}
	if err := ensureTable(diary); err != nil {
		return nil, err
	}
	rows, err := diary.Query(`SELECT id, amount_yuan, alipay_url, alipay_note, wechat_url, wechat_note,
        date_create, IFNULL(date_modify,'') FROM ` + tableName + `
        ORDER BY CASE WHEN amount_yuan = 0 THEN 1 ELSE 0 END, amount_yuan ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.AmountYuan, &item.AlipayURL, &item.AlipayNote, &item.WechatURL, &item.WechatNote, &item.DateCreate, &item.DateModify); err != nil {
			return nil, err
		}
		item.AlipayURL = strings.TrimSpace(item.AlipayURL)
		item.AlipayNote = strings.TrimSpace(item.AlipayNote)
		item.WechatURL = strings.TrimSpace(item.WechatURL)
		item.WechatNote = strings.TrimSpace(item.WechatNote)
		list = append(list, item)
	}
	return list, rows.Err()
}

// HasAmount 列表中是否已有该金额档。
func HasAmount(n int) bool {
	items, err := ListItems()
	if err != nil {
		return false
	}
	for _, item := range items {
		if item.AmountYuan == n {
			return true
		}
	}
	return false
}

func requireAdmin(c *gin.Context) bool {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", errMsg)
		return false
	}
	if !user.IsAdmin() {
		response.Error(c, "", "仅管理员可管理申请收款码")
		return false
	}
	return true
}

func handleList(c *gin.Context) {
	list, err := ListItems()
	if err != nil {
		response.Error(c, err.Error(), "读取收款码列表失败")
		return
	}
	response.Success(c, gin.H{"list": list}, "请求成功")
}

func handleCreate(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var raw map[string]interface{}
	_ = c.ShouldBindJSON(&raw)
	item, err := itemFromMap(raw, 0)
	if err != nil {
		response.Error(c, "", err.Error())
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "保存失败")
		return
	}
	if err := ensureTable(diary); err != nil {
		response.Error(c, err.Error(), "保存失败")
		return
	}
	if exists, err := amountTaken(diary, item.AmountYuan, 0); err != nil {
		response.Error(c, err.Error(), "保存失败")
		return
	} else if exists {
		response.Error(c, "", "该金额已存在")
		return
	}
	now := util.NowString()
	res, err := diary.Exec(`INSERT INTO `+tableName+` (amount_yuan, alipay_url, alipay_note, wechat_url, wechat_note, date_create, date_modify)
        VALUES (?,?,?,?,?,?,?)`, item.AmountYuan, item.AlipayURL, item.AlipayNote, item.WechatURL, item.WechatNote, now, now)
	if err != nil {
		response.Error(c, err.Error(), "保存失败")
		return
	}
	item.ID, _ = res.LastInsertId()
	item.DateCreate = now
	item.DateModify = now
	response.Success(c, item, "已添加")
}

func handleUpdate(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var raw map[string]interface{}
	_ = c.ShouldBindJSON(&raw)
	id := int64(0)
	if n, ok := asInt(raw["id"]); ok {
		id = int64(n)
	}
	if id <= 0 {
		response.Error(c, "", "缺少 id")
		return
	}
	item, err := itemFromMap(raw, id)
	if err != nil {
		response.Error(c, "", err.Error())
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "保存失败")
		return
	}
	if err := ensureTable(diary); err != nil {
		response.Error(c, err.Error(), "保存失败")
		return
	}
	if exists, err := amountTaken(diary, item.AmountYuan, item.ID); err != nil {
		response.Error(c, err.Error(), "保存失败")
		return
	} else if exists {
		response.Error(c, "", "该金额已存在")
		return
	}
	now := util.NowString()
	res, err := diary.Exec(`UPDATE `+tableName+` SET amount_yuan=?, alipay_url=?, alipay_note=?, wechat_url=?, wechat_note=?, date_modify=?
        WHERE id=? LIMIT 1`, item.AmountYuan, item.AlipayURL, item.AlipayNote, item.WechatURL, item.WechatNote, now, item.ID)
	if err != nil {
		response.Error(c, err.Error(), "保存失败")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		response.Error(c, "", "记录不存在")
		return
	}
	item.DateModify = now
	response.Success(c, item, "已保存")
}

func handleDelete(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	var body struct {
		ID int64 `json:"id"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.ID <= 0 {
		response.Error(c, "", "缺少 id")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "删除失败")
		return
	}
	if err := ensureTable(diary); err != nil {
		response.Error(c, err.Error(), "删除失败")
		return
	}
	if _, err := diary.Exec(`DELETE FROM `+tableName+` WHERE id=? LIMIT 1`, body.ID); err != nil {
		response.Error(c, err.Error(), "删除失败")
		return
	}
	removePayQrFiles(body.ID, "alipay")
	removePayQrFiles(body.ID, "wechat")
	response.Success(c, gin.H{"id": body.ID}, "已删除")
}

func handleFileGet(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Status(http.StatusNotFound)
		return
	}
	platform, ok := normalizePlatform(c.Param("platform"))
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}
	abs, contentType, err := findPayQrFile(id, platform)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Cache-Control", "public, max-age=3600")
	c.Header("Content-Type", contentType)
	http.ServeFile(c.Writer, c.Request, abs)
}

func handleUpload(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, "", "缺少 id")
		return
	}
	platform, ok := normalizePlatform(c.Param("platform"))
	if !ok {
		response.Error(c, "", "收款码平台无效，仅支持 alipay / wechat")
		return
	}
	if _, err := getItem(id); err != nil {
		response.Error(c, "", "记录不存在")
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, err.Error(), "未找到上传文件")
		return
	}
	if fileHeader.Size > 4<<20 {
		response.Error(c, "", "收款码图片不能超过 4MB")
		return
	}
	src, err := fileHeader.Open()
	if err != nil {
		response.Error(c, err.Error(), "读取上传文件失败")
		return
	}
	payload, err := io.ReadAll(io.LimitReader(src, 4<<20+1))
	_ = src.Close()
	if err != nil {
		response.Error(c, err.Error(), "读取上传文件失败")
		return
	}
	if int64(len(payload)) > 4<<20 {
		response.Error(c, "", "收款码图片不能超过 4MB")
		return
	}
	contentType := http.DetectContentType(payload)
	ext, ok := payQrContentTypeExt[contentType]
	if !ok {
		response.Error(c, "", "仅支持 png / jpg / webp / gif 图片")
		return
	}
	dir := payQrDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		response.Error(c, err.Error(), "创建收款码目录失败")
		return
	}
	removePayQrFiles(id, platform)
	abs := filepath.Join(dir, fmt.Sprintf("%d-%s%s", id, platform, ext))
	if err := os.WriteFile(abs, payload, 0o644); err != nil {
		response.Error(c, err.Error(), "保存收款码失败")
		return
	}
	publicURL := fmt.Sprintf("%s/%d/%s?v=%d", relPathPrefix, id, platform, time.Now().Unix())
	column := "alipay_url"
	if platform == "wechat" {
		column = "wechat_url"
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "收款码已保存，但写入配置失败")
		return
	}
	if _, err := diary.Exec("UPDATE `"+tableName+"` SET `"+column+"` = ?, date_modify = ? WHERE id = ? LIMIT 1", publicURL, util.NowString(), id); err != nil {
		response.Error(c, err.Error(), "收款码已保存，但写入配置失败")
		return
	}
	item, _ := getItem(id)
	response.Success(c, gin.H{"url": publicURL, "item": item}, "收款码已上传")
}

func mapString(raw map[string]interface{}, key string) string {
	if raw == nil {
		return ""
	}
	if s, ok := raw[key].(string); ok {
		return s
	}
	return ""
}

func itemFromMap(raw map[string]interface{}, id int64) (Item, error) {
	if raw == nil {
		raw = map[string]interface{}{}
	}
	n, ok := asInt(raw["amount_yuan"])
	if !ok || !IsValidAmount(n) {
		return Item{}, fmt.Errorf("请填写 0～100000 的金额，0 表示自定义")
	}
	alipayNote := strings.TrimSpace(mapString(raw, "alipay_note"))
	wechatNote := strings.TrimSpace(mapString(raw, "wechat_note"))
	if utf8.RuneCountInString(alipayNote) > 80 || utf8.RuneCountInString(wechatNote) > 80 {
		return Item{}, fmt.Errorf("备注最多 80 个字")
	}
	return Item{
		ID:         id,
		AmountYuan: n,
		AlipayURL:  strings.TrimSpace(mapString(raw, "alipay_url")),
		AlipayNote: alipayNote,
		WechatURL:  strings.TrimSpace(mapString(raw, "wechat_url")),
		WechatNote: wechatNote,
	}, nil
}

func asInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(n))
		if err == nil {
			return i, true
		}
	}
	return 0, false
}

func amountTaken(diary *sql.DB, amount int, exceptID int64) (bool, error) {
	var id int64
	err := diary.QueryRow(`SELECT id FROM `+tableName+` WHERE amount_yuan = ? AND id <> ? LIMIT 1`, amount, exceptID).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func getItem(id int64) (Item, error) {
	diary, err := diaryDB()
	if err != nil {
		return Item{}, err
	}
	if err := ensureTable(diary); err != nil {
		return Item{}, err
	}
	var item Item
	err = diary.QueryRow(`SELECT id, amount_yuan, alipay_url, alipay_note, wechat_url, wechat_note,
        date_create, IFNULL(date_modify,'') FROM `+tableName+` WHERE id = ? LIMIT 1`, id).
		Scan(&item.ID, &item.AmountYuan, &item.AlipayURL, &item.AlipayNote, &item.WechatURL, &item.WechatNote, &item.DateCreate, &item.DateModify)
	return item, err
}

func findPayQrFile(id int64, platform string) (string, string, error) {
	dir := payQrDir()
	for ct, ext := range payQrContentTypeExt {
		abs := filepath.Join(dir, fmt.Sprintf("%d-%s%s", id, platform, ext))
		if st, err := os.Stat(abs); err == nil && !st.IsDir() {
			return abs, ct, nil
		}
	}
	return "", "", os.ErrNotExist
}

func removePayQrFiles(id int64, platform string) {
	dir := payQrDir()
	seen := map[string]struct{}{}
	for _, ext := range payQrContentTypeExt {
		if _, ok := seen[ext]; ok {
			continue
		}
		seen[ext] = struct{}{}
		_ = os.Remove(filepath.Join(dir, fmt.Sprintf("%d-%s%s", id, platform, ext)))
	}
}

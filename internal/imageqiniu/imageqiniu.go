package imageqiniu

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/systemconfig"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

const dbName = db.Diary
const currentTable = "image_qiniu"

func Register(r *gin.RouterGroup) {
	r.GET("/", handleToken)
	r.GET("/list", handleList)
	r.POST("/add", handleAdd)
	r.DELETE("/delete", handleDelete)
	r.DELETE("/batch-delete", handleBatchDelete)
	r.PUT("/update", handleUpdate)
}

func qiniuKeys() (string, string, error) {
	cfg, err := systemconfig.GetAdminSystemConfig()
	if err != nil {
		return "", "", err
	}
	if cfg.QiniuAccessKey == "" || cfg.QiniuSecretKey == "" {
		return "", "", errors.New("请先在系统配置中填写七牛 Access Key 和 Secret Key")
	}
	return cfg.QiniuAccessKey, cfg.QiniuSecretKey, nil
}

func newMac() (*auth.Credentials, error) {
	ak, sk, err := qiniuKeys()
	if err != nil {
		return nil, err
	}
	return auth.New(ak, sk), nil
}

func uploadToken(bucket string) (string, error) {
	mac, err := newMac()
	if err != nil {
		return "", err
	}
	putPolicy := storage.PutPolicy{Scope: bucket, Expires: 7200}
	return putPolicy.UploadToken(mac), nil
}

func deleteFile(bucket, key string) error {
	mac, err := newMac()
	if err != nil {
		return err
	}
	bm := storage.NewBucketManager(mac, &storage.Config{})
	return bm.Delete(bucket, key)
}

func handleToken(c *gin.Context) {
	bucket := c.Query("bucket")
	if bucket == "" {
		response.Error(c, "", "缺少 bucket 参数")
		return
	}
	if c.Query("hahaha") == "" {
		if _, msg := middleware.VerifyAuthorization(c); msg != "" {
			response.Error(c, "", msg)
			return
		}
	}
	token, err := uploadToken(bucket)
	if err != nil {
		response.Error(c, "", err.Error())
		return
	}
	response.Success(c, token, "凭证获取成功")
}

func handleList(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, msg, "无权查看文件列表：用户信息错误")
		return
	}
	var filters []string
	var args []interface{}

	kws := parseJSONStringArray(c.Query("keywords"))
	if len(kws) > 0 {
		var parts []string
		for _, kw := range kws {
			parts = append(parts, " description like ? ESCAPE '/' ")
			args = append(args, "%"+kw+"%")
		}
		filters = append(filters, strings.Join(parts, " or "))
	}
	if bucket := c.Query("bucket"); bucket != "" {
		filters = append(filters, " bucket = ?")
		args = append(args, bucket)
	}
	whereSQL := ""
	if len(filters) > 0 {
		whereSQL = " where " + strings.Join(filters, " and ")
	}

	pageNo := util.AtoiDefault(c.Query("pageNo"), 1)
	pageSize := util.AtoiDefault(c.Query("pageSize"), 20)
	startPoint := (pageNo - 1) * pageSize

	diary, _ := db.Open(dbName)
	listArgs := append(append([]interface{}{}, args...), startPoint, pageSize)
	list, err := db.QueryMaps(diary, `SELECT * from `+currentTable+whereSQL+` order by date_create desc limit ? , ?`, listArgs...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	countRow, err := db.QueryMap(diary, `select count(*) as sum from `+currentTable+whereSQL, args...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	var total interface{} = 0
	if countRow != nil {
		total = countRow["sum"]
	}
	response.Success(c, gin.H{
		"list":  list,
		"pager": gin.H{"pageSize": pageSize, "pageNo": pageNo, "total": total},
	}, "请求成功")
}

func handleAdd(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Bucket      string `json:"bucket"`
		BaseURL     string `json:"base_url"`
	}
	_ = c.ShouldBindJSON(&body)
	now := util.NowString()
	diary, _ := db.Open(dbName)
	_, err := diary.Exec(`insert into `+currentTable+`(id, description, date_create, type, bucket, base_url) values(?,?,?,?,?,?)`,
		body.ID, body.Description, now, body.Type, body.Bucket, body.BaseURL)
	if err != nil {
		response.Error(c, err.Error(), "七牛云图片添加失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": body.ID}, "添加成功")
}

func handleDelete(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		ID string `json:"id"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	fileInfo, err := db.QueryMap(diary, `SELECT * from `+currentTable+` where id=?`, body.ID)
	if err != nil {
		response.Error(c, "", err.Error())
		return
	}
	if fileInfo == nil {
		response.Error(c, "", "文件不存在")
		return
	}
	if err := deleteFile(asString(fileInfo["bucket"]), asString(fileInfo["id"])); err != nil {
		response.Error(c, "", err.Error())
		return
	}
	_, err = diary.Exec(`delete from `+currentTable+` where id = ?`, body.ID)
	if err != nil {
		response.Error(c, err.Error(), "七牛云图片删除失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "删除成功")
}

func handleBatchDelete(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		IDs    []string `json:"ids"`
		Bucket string   `json:"bucket"`
	}
	_ = c.ShouldBindJSON(&body)
	if len(body.IDs) == 0 {
		response.Error(c, "", "无效的删除请求")
		return
	}
	diary, _ := db.Open(dbName)
	placeholders := make([]string, len(body.IDs))
	args := make([]interface{}, len(body.IDs))
	for i, id := range body.IDs {
		placeholders[i] = "?"
		args[i] = id
	}
	fileInfos, err := db.QueryMaps(diary, `SELECT * from `+currentTable+` where id in (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		response.Error(c, "", err.Error())
		return
	}
	if len(fileInfos) == 0 {
		response.Error(c, "", "文件不存在")
		return
	}
	for _, f := range fileInfos {
		if err := deleteFile(asString(f["bucket"]), asString(f["id"])); err != nil {
			response.Error(c, "", err.Error())
			return
		}
	}
	_, err = diary.Exec(`delete from `+currentTable+` where id in (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		response.Error(c, err.Error(), "七牛云图片批量删除失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "批量删除成功")
}

func handleUpdate(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		ID          string `json:"id"`
		Description string `json:"description"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.ID == "" {
		response.Error(c, "", "参数错误")
		return
	}
	diary, _ := db.Open(dbName)
	_, err := diary.Exec(`update `+currentTable+` set description = ? where id = ?`, body.Description, body.ID)
	if err != nil {
		response.Error(c, err.Error(), "七牛云图片更新失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "更新成功")
}

func parseJSONStringArray(s string) []string {
	if s == "" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil
	}
	return arr
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return ""
	}
}

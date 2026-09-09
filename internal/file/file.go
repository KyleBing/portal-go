package file

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/setup"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const dbName = db.Diary
const currentTable = "file_manager"
const destFolder = "upload"

func Register(r *gin.RouterGroup) {
	r.POST("/upload", handleUpload)
	r.POST("/modify", handleModify)
	r.DELETE("/delete", handleDelete)
	r.GET("/list", handleList)
	r.GET("/download", handleDownload)
}

func uploadDir() string {
	return filepath.Join(setup.ProjectRoot(), destFolder)
}

func absUnderUpload(rel string) (string, error) {
	rel = filepath.Clean(rel)
	if rel == "." || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("非法路径")
	}
	abs := filepath.Join(setup.ProjectRoot(), rel)
	root := uploadDir()
	if abs != root && !strings.HasPrefix(abs, root+string(filepath.Separator)) {
		return "", fmt.Errorf("非法路径")
	}
	return abs, nil
}

func handleUpload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, err.Error(), "未找到上传文件")
		return
	}
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, msg, "无权操作")
		return
	}
	name := filepath.Base(fileHeader.Filename)
	if name == "." || name == ".." || name == "" || strings.Contains(name, "\x00") {
		response.Error(c, "", "非法文件名")
		return
	}

	userDirRel := filepath.Join(destFolder, strconv.FormatInt(user.UID, 10))
	userDirAbs := filepath.Join(uploadDir(), strconv.FormatInt(user.UID, 10))
	if err := os.MkdirAll(userDirAbs, 0o755); err != nil {
		response.Error(c, err.Error(), "上传失败")
		return
	}

	destPath := filepath.ToSlash(filepath.Join(userDirRel, name))
	absDest := filepath.Join(userDirAbs, name)
	if _, err := absUnderUpload(destPath); err != nil {
		response.Error(c, "", "非法文件路径")
		return
	}
	if _, statErr := os.Stat(absDest); statErr == nil {
		response.Error(c, "", "文件已存在")
		return
	}

	diary, _ := db.Open(dbName)
	tx, err := diary.Begin()
	if err != nil {
		response.Error(c, err.Error(), "beginTransaction: 事务执行失败，已回滚")
		return
	}
	mimeType := fileHeader.Header.Get("Content-Type")
	now := util.NowString()
	note := c.PostForm("note")
	res, err := tx.Exec(`insert into `+currentTable+`(path, name_original, description, date_create, type, size, uid) values (?,?,?,?,?,?,?)`,
		destPath, name, note, now, mimeType, fileHeader.Size, user.UID)
	if err != nil {
		_ = tx.Rollback()
		response.Error(c, err.Error(), "query: sql 事务执行失败，已回滚")
		return
	}
	if err := c.SaveUploadedFile(fileHeader, absDest); err != nil {
		_ = tx.Rollback()
		response.Error(c, err.Error(), "上传失败")
		return
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		_ = os.Remove(absDest)
		response.Error(c, err.Error(), "transaction.commit: 事务执行失败，已回滚")
		return
	}
	id, _ := res.LastInsertId()
	response.Success(c, gin.H{
		"id":           id,
		"path":         destPath,
		"download_url": fmt.Sprintf("/portal/file-manager/download?fileId=%d", id),
	}, "上传成功")
}

func handleModify(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body struct {
		FileID      int64  `json:"fileId"`
		Description string `json:"description"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	_, err := diary.Exec(`update `+currentTable+` set description = ? WHERE id=? and uid=?`, body.Description, body.FileID, user.UID)
	if err != nil {
		response.Error(c, err.Error(), "文件修改失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "修改成功")
}

func handleDelete(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body struct {
		FileID int64 `json:"fileId"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	fileInfo, err := db.QueryMap(diary, `select * from `+currentTable+` where id=? and uid=?`, body.FileID, user.UID)
	if err != nil {
		response.Error(c, "", err.Error())
		return
	}
	res, err := diary.Exec(`DELETE from `+currentTable+` WHERE id=? and uid=?`, body.FileID, user.UID)
	if err != nil {
		response.Error(c, err.Error(), "")
		return
	}
	affected, _ := res.RowsAffected()
	if affected > 0 {
		util.UpdateUserLastLoginTime(user.UID)
		if fileInfo != nil {
			p := asString(fileInfo["path"])
			if p != "" {
				if abs, err := absUnderUpload(p); err == nil {
					_ = os.Remove(abs)
				}
			}
		}
		response.Success(c, "", "删除成功")
	} else {
		response.Error(c, "", "删除失败")
	}
}

func handleList(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	pageNo := util.AtoiDefault(c.Query("pageNo"), 1)
	pageSize := util.AtoiDefault(c.Query("pageSize"), 20)
	startPoint := (pageNo - 1) * pageSize

	var sqlBuf string
	args := []interface{}{user.UID}

	for _, kw := range parseJSONStringArray(c.Query("keywords")) {
		sqlBuf += ` and ( description like ? ESCAPE '/' )`
		args = append(args, "%"+util.UnicodeEncode(kw)+"%")
	}
	if df := c.Query("dateFilter"); len(df) >= 6 {
		year := df[0:4]
		month := df[4:6]
		sqlBuf += ` and YEAR(date_create)=? AND MONTH(date_create)=?`
		args = append(args, year, month)
	}
	args = append(args, startPoint, pageSize)

	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, `SELECT * from `+currentTable+` where uid=?`+sqlBuf+` order by date_create desc limit ?, ?`, args...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	for _, row := range rows {
		id := asInt64(row["id"])
		row["download_url"] = fmt.Sprintf("/portal/file-manager/download?fileId=%d", id)
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, rows, "请求成功")
}

func handleDownload(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	fileID, err := strconv.ParseInt(c.Query("fileId"), 10, 64)
	if err != nil || fileID <= 0 {
		response.Error(c, "", "参数错误")
		return
	}
	diary, _ := db.Open(dbName)
	row, err := db.QueryMap(diary, `select * from `+currentTable+` where id=? and uid=?`, fileID, user.UID)
	if err != nil || row == nil {
		response.Error(c, "", "文件不存在或无权访问")
		return
	}
	rel := asString(row["path"])
	abs, err := absUnderUpload(rel)
	if err != nil {
		response.Error(c, "", "非法文件路径")
		return
	}
	if _, err := os.Stat(abs); err != nil {
		response.Error(c, "", "磁盘文件不存在")
		return
	}
	name := asString(row["name_original"])
	if name == "" {
		name = filepath.Base(abs)
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, strings.ReplaceAll(name, `"`, ``)))
	http.ServeFile(c.Writer, c.Request, abs)
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
		return fmt.Sprintf("%v", s)
	}
}

func asInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case []byte:
		x, _ := strconv.ParseInt(string(n), 10, 64)
		return x
	case string:
		x, _ := strconv.ParseInt(n, 10, 64)
		return x
	}
	return 0
}

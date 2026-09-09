package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
}

func uploadDir() string {
	return filepath.Join(setup.ProjectRoot(), destFolder)
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
	destPath := destFolder + "/" + name
	absDest := filepath.Join(uploadDir(), name)
	// 防止路径穿越：落盘路径必须仍在 upload 目录内
	if !strings.HasPrefix(absDest, uploadDir()+string(filepath.Separator)) && absDest != uploadDir() {
		response.Error(c, "", "非法文件路径")
		return
	}

	if _, statErr := os.Stat(absDest); statErr == nil {
		response.Error(c, "", "文件已存在")
		return
	}
	if err := os.MkdirAll(uploadDir(), 0o755); err != nil {
		response.Error(c, err.Error(), "上传失败")
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
	_, err = tx.Exec(`insert into `+currentTable+`(path, name_original, description, date_create, type, size, uid) values (?,?,?,?,?,?,?)`,
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
	response.Success(c, "", "上传成功")
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
	fileInfo, err := db.QueryMap(diary, `select * from `+currentTable+` where id=?`, body.FileID)
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
				_ = os.Remove(filepath.Join(setup.ProjectRoot(), p))
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
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, rows, "请求成功")
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

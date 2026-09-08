package invitation

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/KyleBing/portal-go/internal/apihelper"
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const dataName = "邀请码"
const table = "invitations"

// Register mounts /invitation routes.
func Register(r *gin.RouterGroup) {
	g := r
	g.GET("/list", handleList)
	g.POST("/generate", handleGenerate)
	g.POST("/mark-shared", handleMarkShared)
	g.DELETE("/delete", handleDelete)
}

func handleList(c *gin.Context) {
	diary, err := db.Open(db.Diary)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		// unauthenticated: only shared, unbound codes
		data, e := apihelper.QueryMaps(diary, `SELECT * from `+table+` where binding_uid is null and is_shared = 0 order by date_create desc`)
		if e != nil {
			response.Error(c, e.Error(), e.Error())
			return
		}
		response.Success(c, data, "请求成功")
		return
	}
	var data []map[string]interface{}
	if user.IsAdmin() {
		data, err = apihelper.QueryMaps(diary, `SELECT * from `+table+` where binding_uid is null order by date_create desc`)
	} else {
		data, err = apihelper.QueryMaps(diary, `SELECT * from `+table+` where binding_uid is null and is_shared = 0 order by date_create desc`)
	}
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, data, "请求成功")
}

func handleGenerate(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, errMsg, errMsg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权限操作")
		return
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		response.Error(c, err.Error(), dataName+"生成失败")
		return
	}
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		response.Error(c, err.Error(), dataName+"生成失败")
		return
	}
	key := base64.StdEncoding.EncodeToString(buf)
	now := util.NowString()
	apihelper.OperateReturnID(c, diary, user.UID, dataName, "生成", `insert into `+table+`(date_create, id) VALUES (?, ?)`, now, key)
}

func handleMarkShared(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, errMsg, errMsg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权限操作")
		return
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		response.Error(c, err.Error(), dataName+"标记失败")
		return
	}
	body := apihelper.Body(c)
	apihelper.OperateReturnID(c, diary, user.UID, dataName, "标记", `update `+table+` set is_shared = 1 where id = ?`, apihelper.S(body, "id"))
}

func handleDelete(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		response.Error(c, "", "参数错误，缺少 id")
		return
	}
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, errMsg, "无权操作")
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		response.Error(c, err.Error(), dataName+"添加失败")
		return
	}
	apihelper.OperateReturnID(c, diary, user.UID, dataName, "添加", `DELETE from `+table+` WHERE id=?`, id)
}

package category

import (
	"github.com/KyleBing/portal-go/internal/apihelper"
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const dataName = "日记类别"
const table = "diary_category"

// Register mounts /diary-category routes.
func Register(r *gin.RouterGroup) {
	g := r
	g.GET("/list", handleList)
	g.POST("/add", handleAdd)
	g.PUT("/modify", handleModify)
	g.DELETE("/delete", handleDelete)
}

func handleList(c *gin.Context) {
	diary, err := db.Open(db.Diary)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	data, err := apihelper.QueryMaps(diary, `select * from `+table+` order by sort_id asc`)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	response.Success(c, data, "")
}

func handleAdd(c *gin.Context) {
	diary, err := db.Open(db.Diary)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	body := apihelper.Body(c)
	exist, err := apihelper.QueryMaps(diary, `select * from `+table+` where name_en=?`, apihelper.S(body, "name_en"))
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	if len(exist) > 0 {
		response.Error(c, "", dataName+"已存在")
		return
	}
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", errMsg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	now := util.NowString()
	apihelper.OperateReturnID(c, diary, user.UID, dataName, "添加",
		`insert into `+table+`(name, name_en, color, sort_id, date_init) values(?,?,?,?,?)`,
		apihelper.S(body, "name"), apihelper.S(body, "name_en"), apihelper.S(body, "color"), apihelper.S(body, "sort_id"), now)
}

func handleModify(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", errMsg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		response.Error(c, err.Error(), dataName+"修改失败")
		return
	}
	body := apihelper.Body(c)
	apihelper.OperateNoReturn(c, diary, user.UID, dataName, "修改",
		`update `+table+` set name=?, count=?, color=?, sort_id=? where name_en=?`,
		apihelper.S(body, "name"), apihelper.S(body, "count"), apihelper.S(body, "color"), apihelper.I(body, "sort_id"), apihelper.S(body, "name_en"))
}

func handleDelete(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", errMsg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		response.Error(c, err.Error(), dataName+"删除失败")
		return
	}
	body := apihelper.Body(c)
	apihelper.OperateNoReturn(c, diary, user.UID, dataName, "删除", `delete from `+table+` where name_en = ?`, apihelper.S(body, "name_en"))
}

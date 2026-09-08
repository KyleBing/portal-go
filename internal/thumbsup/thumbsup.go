package thumbsup

import (
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const dbName = db.Diary
const currentTable = "thumbs_up"

func Register(r *gin.RouterGroup) {
	r.GET("/", handleGet)
	r.GET("/all", handleAll)
	r.GET("/list", handleList)
	r.POST("/add", handleAdd)
	r.PUT("/modify", handleModify)
	r.DELETE("/delete", handleDelete)
}

func handleGet(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		response.Error(c, "", "key 值未定义")
		return
	}
	diary, _ := db.Open(dbName)
	row, err := db.QueryMap(diary, `SELECT * from `+currentTable+` where name = ?`, key)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	var count interface{}
	if row != nil {
		count = row["count"]
	}
	response.Success(c, count, "请求成功")
}

func handleAll(c *gin.Context) {
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, `SELECT * from `+currentTable)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	response.Success(c, rows, "请求成功")
}

func handleList(c *gin.Context) {
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, `select * from `+currentTable+` order by date_init asc`)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	response.Success(c, rows, "")
}

func handleAdd(c *gin.Context) {
	var body struct {
		Name        string `json:"name"`
		Count       int    `json:"count"`
		Description string `json:"description"`
		LinkAddress string `json:"link_address"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	exist, _ := db.QueryMaps(diary, `select * from `+currentTable+` where name=?`, body.Name)
	if len(exist) > 0 {
		response.Error(c, "", "点赞信息名已存在")
		return
	}
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	now := util.NowString()
	res, err := diary.Exec(`insert into `+currentTable+`(name, count, description, link_address, date_init) values(?,?,?,?,?)`,
		body.Name, body.Count, body.Description, body.LinkAddress, now)
	if err != nil {
		response.Error(c, err.Error(), "点赞添加失败")
		return
	}
	id, _ := res.LastInsertId()
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": id}, "添加成功")
}

func handleModify(c *gin.Context) {
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
		Name        string `json:"name"`
		Count       int    `json:"count"`
		Description string `json:"description"`
		LinkAddress string `json:"link_address"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	_, err := diary.Exec(`update `+currentTable+` set name = ?, count = ?, description = ?, link_address = ? where name = ?`,
		body.Name, body.Count, body.Description, body.LinkAddress, body.Name)
	if err != nil {
		response.Error(c, err.Error(), "点赞编辑失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "编辑成功")
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
		Name string `json:"name"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	_, err := diary.Exec(`delete from `+currentTable+` where name = ?`, body.Name)
	if err != nil {
		response.Error(c, err.Error(), "点赞删除失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "删除成功")
}

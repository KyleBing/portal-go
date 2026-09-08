package mappointer

import (
	"fmt"
	"strings"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/models"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const dbName = db.Diary
const currentTable = "map_pointer"

const selectBase = `select
	map_pointer.id, map_pointer.name, map_pointer.pointers, map_pointer.note, map_pointer.area,
	map_pointer.date_create, map_pointer.date_modify, map_pointer.thumb_up, map_pointer.is_public, map_pointer.uid,
	users.wx, users.uid, users.nickname, users.username
	from map_pointer left join users on map_pointer.uid = users.uid`

func Register(r *gin.RouterGroup) {
	r.POST("/list", handleList)
	r.GET("/detail", handleDetail)
	r.POST("/add", handleAdd)
	r.PUT("/modify", handleModify)
	r.DELETE("/delete", handleDelete)
}

type listBody struct {
	Keyword   string   `json:"keyword"`
	DateRange []string `json:"dateRange"`
	PageNo    int      `json:"pageNo"`
	PageSize  int      `json:"pageSize"`
}

func handleList(c *gin.Context) {
	var body listBody
	_ = c.ShouldBindJSON(&body)
	user, _ := middleware.VerifyAuthorization(c)
	getPointerList(user, body, c)
}

func getPointerList(user *models.User, body listBody, c *gin.Context) {
	var filters []string
	var args []interface{}

	if user != nil {
		filters = append(filters, "(is_public = 1 or map_pointer.uid = ?)")
		args = append(args, user.UID)
	} else {
		filters = append(filters, "is_public = 1")
	}

	if body.Keyword != "" {
		var kwParts []string
		for _, kw := range strings.Split(body.Keyword, " ") {
			kwParts = append(kwParts, " ( map_pointer.name like ? ESCAPE '/'  or  map_pointer.note like ? ESCAPE '/') ")
			like := "%" + util.UnicodeEncode(kw) + "%"
			args = append(args, like, like)
		}
		if len(kwParts) > 0 {
			filters = append(filters, strings.Join(kwParts, " and "))
		}
	}

	if len(body.DateRange) == 2 {
		filters = append(filters, "date_init between ? AND ?")
		args = append(args, body.DateRange[0], body.DateRange[1])
	}

	filterSQL := ""
	if len(filters) > 0 {
		filterSQL = "where " + strings.Join(filters, " and ")
	}

	pointStart := (body.PageNo - 1) * body.PageSize
	diary, _ := db.Open(dbName)

	listArgs := append(append([]interface{}{}, args...), pointStart, body.PageSize)
	list, err := db.QueryMaps(diary, selectBase+" "+filterSQL+" limit ? , ?", listArgs...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	countRow, err := db.QueryMap(diary, "select count(*) as sum from "+currentTable+" "+filterSQL, args...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	for _, item := range list {
		item["name"] = util.UnicodeDecode(asString(item["name"]))
		item["note"] = util.UnicodeDecode(asString(item["note"]))
	}
	var total interface{} = 0
	if countRow != nil {
		total = countRow["sum"]
	}
	response.Success(c, gin.H{
		"list": list,
		"pager": gin.H{
			"pageSize": body.PageSize,
			"pageNo":   body.PageNo,
			"total":    total,
		},
	}, "请求成功")
}

func handleDetail(c *gin.Context) {
	id := c.Query("id")
	diary, _ := db.Open(dbName)
	row, err := db.QueryMap(diary, selectBase+" where map_pointer.id = ?", id)
	if err != nil || row == nil {
		response.Error(c, errStr(err), "查无此路线")
		return
	}
	if asInt(row["is_public"]) == 1 {
		response.Success(c, row, "")
		return
	}
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	if asInt(row["uid"]) == user.UID || user.IsAdmin() {
		response.Success(c, row, "")
	} else {
		response.Error(c, "", "该路线信息不属于您，无权操作")
	}
}

func handleAdd(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body struct {
		Name     string `json:"name"`
		Pointers string `json:"pointers"`
		Note     string `json:"note"`
		Area     string `json:"area"`
		ThumbUp  int    `json:"thumb_up"`
		IsPublic int    `json:"is_public"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	encodedName := util.UnicodeEncode(body.Name)
	// 与原实现保持一致：查询的是 map_route 表
	exist, _ := db.QueryMaps(diary, `select * from map_route where name=?`, encodedName)
	if len(exist) > 0 {
		response.Error(c, "", fmt.Sprintf("已存在名为 %s 的地域信息", body.Name))
		return
	}
	now := util.NowString()
	res, err := diary.Exec(`insert into `+currentTable+`(name, pointers, note, uid, date_create, date_modify, area, thumb_up, is_public)
		values(?,?,?,?,?,?,?,?,?)`,
		encodedName, body.Pointers, util.UnicodeEncode(body.Note), user.UID, now, now, body.Area, body.ThumbUp, body.IsPublic)
	if err != nil {
		response.Error(c, err.Error(), "路书标记点添加失败")
		return
	}
	id, _ := res.LastInsertId()
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": id}, "添加成功")
}

func handleModify(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "查无此路线")
		return
	}
	var body struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Pointers string `json:"pointers"`
		Note     string `json:"note"`
		Area     string `json:"area"`
		IsPublic int    `json:"is_public"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	row, err := db.QueryMap(diary, `select * from `+currentTable+` where id = ?`, body.ID)
	if err != nil || row == nil {
		response.Error(c, errStr(err), "查无此路线")
		return
	}
	if asInt(row["uid"]) != user.UID && !user.IsAdmin() {
		response.Error(c, "", "该路线信息不属于您，无权操作")
		return
	}
	now := util.NowString()
	res, err := diary.Exec(`update `+currentTable+` set name=?, pointers=?, note=?, date_modify=?, area=?, is_public=? WHERE id=?`,
		util.UnicodeEncode(body.Name), body.Pointers, util.UnicodeEncode(body.Note), now, body.Area, body.IsPublic, body.ID)
	if err != nil {
		response.Error(c, err.Error(), "路书标记点修改失败")
		return
	}
	id, _ := res.LastInsertId()
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": id}, "修改成功")
}

func handleDelete(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "查无此路线")
		return
	}
	var body struct {
		ID int64 `json:"id"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	row, err := db.QueryMap(diary, `select * from `+currentTable+` where id = ?`, body.ID)
	if err != nil || row == nil {
		response.Error(c, errStr(err), "查无此路线")
		return
	}
	if asInt(row["uid"]) != user.UID && !user.IsAdmin() {
		response.Error(c, "", "该路线不属于您，无权操作")
		return
	}
	_, err = diary.Exec(`DELETE from `+currentTable+` WHERE id=?`, body.ID)
	if err != nil {
		response.Error(c, err.Error(), "路书标记点删除失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "删除成功")
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

func asInt(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case []byte:
		var x int64
		fmt.Sscan(string(n), &x)
		return x
	case string:
		var x int64
		fmt.Sscan(n, &x)
		return x
	}
	return 0
}

func errStr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

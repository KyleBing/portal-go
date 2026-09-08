package maproute

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
const currentTable = "map_route"

const selectBase = `select
	map_route.id, map_route.name, map_route.area, map_route.road_type, map_route.seasons,
	map_route.video_link, map_route.paths, map_route.note, map_route.date_init, map_route.date_modify,
	map_route.thumb_up, map_route.is_public, map_route.policy, map_route.uid,
	users.phone, users.wx, users.uid, users.nickname, users.username
	from map_route left join users on map_route.uid = users.uid`

func Register(r *gin.RouterGroup) {
	r.POST("/list", handleList)
	r.GET("/detail", handleDetail)
	r.POST("/add", handleAdd)
	r.PUT("/modify", handleModify)
	r.DELETE("/delete", handleDelete)
}

type listBody struct {
	IsMine    string   `json:"isMine"`
	Keyword   string   `json:"keyword"`
	DateRange []string `json:"dateRange"`
	PageNo    int      `json:"pageNo"`
	PageSize  int      `json:"pageSize"`
}

func handleList(c *gin.Context) {
	var body listBody
	_ = c.ShouldBindJSON(&body)
	user, _ := middleware.VerifyAuthorization(c)
	getRouteLineList(user, body, c)
}

func getRouteLineList(user *models.User, body listBody, c *gin.Context) {
	var filters []string
	var args []interface{}

	if user != nil {
		if body.IsMine == "1" {
			filters = append(filters, "map_route.uid = ?")
			args = append(args, user.UID)
		} else if user.IsAdmin() {
			filters = append(filters, "map_route.uid != ?")
			args = append(args, user.UID)
		} else {
			filters = append(filters, "is_public = 1 and map_route.uid != ?")
			args = append(args, user.UID)
		}
	} else {
		filters = append(filters, "is_public = 1")
	}

	if body.Keyword != "" {
		var kwParts []string
		for _, kw := range strings.Split(body.Keyword, " ") {
			kwParts = append(kwParts, " (map_route.name like ? ESCAPE '/')")
			args = append(args, "%"+util.UnicodeEncode(kw)+"%")
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
	row, err := db.QueryMap(diary, selectBase+" where map_route.id = ?", id)
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
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		Name      string      `json:"name"`
		Area      string      `json:"area"`
		RoadType  string      `json:"road_type"`
		Policy    interface{} `json:"policy"`
		Seasons   string      `json:"seasons"`
		VideoLink string      `json:"video_link"`
		Paths     string      `json:"paths"`
		Note      string      `json:"note"`
		ThumbUp   int         `json:"thumb_up"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	encodedName := util.UnicodeEncode(body.Name)
	exist, _ := db.QueryMaps(diary, `select * from map_route where name=?`, encodedName)
	if len(exist) > 0 {
		response.Error(c, "", fmt.Sprintf("已存在名为 %s 的路线", body.Name))
		return
	}
	now := util.NowString()
	res, err := diary.Exec(`insert into `+currentTable+`(name, area, road_type, policy, seasons, video_link, paths, note, date_init, date_modify, thumb_up, uid)
		values(?,?,?,?,?,?,?,?,?,?,?,?)`,
		encodedName, body.Area, body.RoadType, body.Policy, body.Seasons, body.VideoLink, body.Paths,
		util.UnicodeEncode(body.Note), now, now, body.ThumbUp, user.UID)
	if err != nil {
		response.Error(c, err.Error(), "路书路径添加失败")
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
		ID        int64       `json:"id"`
		Name      string      `json:"name"`
		Area      string      `json:"area"`
		RoadType  string      `json:"road_type"`
		Policy    interface{} `json:"policy"`
		Seasons   string      `json:"seasons"`
		VideoLink string      `json:"video_link"`
		Paths     string      `json:"paths"`
		Note      string      `json:"note"`
		IsPublic  int         `json:"is_public"`
		ThumbUp   int         `json:"thumb_up"`
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
	res, err := diary.Exec(`update `+currentTable+` set name=?, area=?, road_type=?, policy=?, seasons=?, video_link=?, paths=?, note=?, date_modify=?, is_public=?, thumb_up=? WHERE id=?`,
		util.UnicodeEncode(body.Name), body.Area, body.RoadType, body.Policy, body.Seasons, body.VideoLink,
		body.Paths, util.UnicodeEncode(body.Note), now, body.IsPublic, body.ThumbUp, body.ID)
	if err != nil {
		response.Error(c, err.Error(), "路书路径修改失败")
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
		response.Error(c, err.Error(), "路书路径删除失败")
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

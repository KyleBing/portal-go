package diary

import (
	"encoding/json"
	"fmt"

	"github.com/KyleBing/portal-go/internal/bill"
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/systemconfig"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const dbName = db.Diary

// EnumUserGroupAdmin 管理员组别
const enumUserGroupAdmin = 1

func Register(r *gin.RouterGroup) {
	r.GET("/list", handleList)
	r.GET("/list-all", handleListAll)
	r.GET("/list-title-only", handleListTitleOnly)
	r.GET("/list-category-only", handleListCategoryOnly)
	r.GET("/export", handleExport)
	r.GET("/temperature", handleTemperature)
	r.GET("/detail", handleDetail)
	r.POST("/add", handleAdd)
	r.PUT("/modify", handleModify)
	r.DELETE("/delete", handleDelete)
	r.GET("/latest-recommend", handleLatestRecommend)
	r.GET("/get-diary-content-with-keyword", handleGetContentWithKeyword)
	r.GET("/get-latest-public-diary-with-keyword", handleGetLatestPublicWithKeyword)
	r.POST("/clear", handleClear)
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

// buildFilter 追加日记查询的过滤条件（在 `where uid=?` 之后），返回 SQL 片段和参数
func buildFilter(c *gin.Context) (string, []interface{}) {
	var sqlBuf string
	var args []interface{}

	// keywords
	for _, kw := range parseJSONStringArray(c.Query("keywords")) {
		encoded := util.UnicodeEncode(kw)
		sqlBuf += ` and ( title like ? ESCAPE '/'  or content like ? ESCAPE '/')`
		like := "%" + encoded + "%"
		args = append(args, like, like)
	}

	// categories
	cats := parseJSONStringArray(c.Query("categories"))
	if len(cats) > 0 {
		sqlBuf += ` and (`
		for i, cat := range cats {
			if i > 0 {
				sqlBuf += ` or `
			}
			sqlBuf += `category=?`
			args = append(args, cat)
		}
		sqlBuf += `)`
	}

	// share
	if c.Query("filterShared") == "1" {
		sqlBuf += ` and is_public = 1`
	}

	// time range
	timeStart := c.Query("timeStart")
	timeEnd := c.Query("timeEnd")
	if timeStart != "" && timeEnd != "" {
		sqlBuf += ` and date >= ? and date <= ?`
		args = append(args, timeStart, timeEnd)
	} else if timeStart != "" {
		sqlBuf += ` and date >= ?`
		args = append(args, timeStart)
	} else if timeEnd != "" {
		sqlBuf += ` and date <= ?`
		args = append(args, timeEnd)
	}

	return sqlBuf, args
}

// decodeDiaryRow 处理日记标题、内容的 unicode 解码及账单数据。
// Unescape runs before UnicodeDecode so MySQL-style \\ collapses before \uXXXX recovery.
func decodeDiaryRow(row map[string]interface{}, unescape bool, withBill bool) {
	if t, ok := row["title"]; ok {
		title := asString(t)
		if unescape {
			title = util.UnescapeMySQLString(title)
		}
		row["title"] = util.UnicodeDecode(title)
	}
	if ct, ok := row["content"]; ok {
		content := asString(ct)
		if unescape {
			content = util.UnescapeMySQLString(content)
		}
		row["content"] = util.UnicodeDecode(content)
	}
	if withBill && asString(row["category"]) == "bill" {
		row["billData"] = bill.ProcessBillOfDay(map[string]interface{}{
			"id":       row["id"],
			"month_id": row["month_id"],
			"date":     row["date"],
			"content":  util.UnicodeEncode(asString(row["content"])),
		}, nil)
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

	filterSQL, args := buildFilter(c)
	query := `SELECT * from diaries where uid=?` + filterSQL + ` order by date desc limit ?, ?`
	allArgs := append([]interface{}{user.UID}, args...)
	allArgs = append(allArgs, startPoint, pageSize)

	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, query, allArgs...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	for _, row := range rows {
		decodeDiaryRow(row, true, true)
	}
	response.Success(c, rows, "请求成功")
}

func handleListAll(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	filterSQL, args := buildFilter(c)
	query := `SELECT * from diaries where uid=?` + filterSQL + ` order by date desc`
	allArgs := append([]interface{}{user.UID}, args...)
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, query, allArgs...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	for _, row := range rows {
		decodeDiaryRow(row, true, true)
	}
	response.Success(c, rows, "请求成功")
}

func handleListTitleOnly(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	filterSQL, args := buildFilter(c)
	query := `SELECT id,date,title,category from diaries where uid=?` + filterSQL + ` order by date desc`
	allArgs := append([]interface{}{user.UID}, args...)
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, query, allArgs...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	for _, row := range rows {
		if t, ok := row["title"]; ok {
			row["title"] = util.UnicodeDecode(asString(t))
		}
	}
	response.Success(c, rows, "请求成功")
}

func handleListCategoryOnly(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	filterSQL, args := buildFilter(c)
	query := `SELECT id,DATE_FORMAT(date,'%Y-%m-%d') as date,category from diaries where uid=?` + filterSQL + ` order by date desc`
	allArgs := append([]interface{}{user.UID}, args...)
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, query, allArgs...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, rows, "请求成功")
}

func handleExport(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权查看日记列表：用户信息错误")
		return
	}
	filterSQL, args := buildFilter(c)
	query := `SELECT * from diaries where uid=?` + filterSQL + ` order by date desc`
	allArgs := append([]interface{}{user.UID}, args...)
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, query, allArgs...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	for _, row := range rows {
		decodeDiaryRow(row, true, true)
	}
	response.Success(c, rows, "请求成功")
}

func handleTemperature(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var sqlBuf string
	var args []interface{}
	args = append(args, user.UID)
	timeStart := c.Query("timeStart")
	timeEnd := c.Query("timeEnd")
	if timeStart != "" && timeEnd != "" {
		sqlBuf += ` and date >= ? and date <= ?`
		args = append(args, timeStart, timeEnd)
	} else if timeStart != "" {
		sqlBuf += ` and date >= ?`
		args = append(args, timeStart)
	} else if timeEnd != "" {
		sqlBuf += ` and date <= ?`
		args = append(args, timeEnd)
	}
	query := `SELECT date, temperature, temperature_outside FROM diaries WHERE uid=? AND category = 'life'` + sqlBuf + ` order by date desc limit 100`
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, query, args...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	for _, row := range rows {
		decodeDiaryRow(row, true, false)
	}
	response.Success(c, rows, "请求成功")
}

func handleDetail(c *gin.Context) {
	diaryID := c.Query("diaryId")
	diary, _ := db.Open(dbName)
	row, err := db.QueryMap(diary, `select * from diaries where id = ?`, diaryID)
	if err != nil {
		response.Error(c, err.Error(), "")
		return
	}
	if row == nil {
		response.Error(c, "", "")
		return
	}
	decodeDiaryRow(row, true, false)
	if asInt(row["is_public"]) == 1 {
		ownerID := asInt(row["uid"])
		u, _ := db.QueryMap(diary, `select * from users where uid = ?`, ownerID)
		if u != nil {
			row["nickname"] = u["nickname"]
			row["username"] = u["username"]
		}
		response.Success(c, row, "")
		return
	}
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	if user.UID == asInt(row["uid"]) {
		util.UpdateUserLastLoginTime(user.UID)
		response.Success(c, row, "")
	} else {
		response.Error(c, "", "无权查看该日记：请求用户 ID 与日记归属不匹配")
	}
}

func handleAdd(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body struct {
		Title              string  `json:"title"`
		Content            string  `json:"content"`
		Category           string  `json:"category"`
		Weather            string  `json:"weather"`
		Temperature        float64 `json:"temperature"`
		TemperatureOutside float64 `json:"temperature_outside"`
		Date               string  `json:"date"`
		IsPublic           int     `json:"is_public"`
		IsMarkdown         int     `json:"is_markdown"`
	}
	_ = c.ShouldBindJSON(&body)
	now := util.NowString()
	diary, _ := db.Open(dbName)
	res, err := diary.Exec(`INSERT into diaries(title, content, category, weather, temperature, temperature_outside, date_create, date_modify, date, uid, is_public, is_markdown)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		util.UnicodeEncode(body.Title), util.UnicodeEncode(body.Content), body.Category, body.Weather,
		body.Temperature, body.TemperatureOutside, now, now, body.Date, user.UID, body.IsPublic, body.IsMarkdown)
	if err != nil {
		response.Error(c, err.Error(), "日记添加失败")
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
	var body struct {
		ID                 int64   `json:"id"`
		Title              string  `json:"title"`
		Content            string  `json:"content"`
		Category           string  `json:"category"`
		Weather            string  `json:"weather"`
		Temperature        float64 `json:"temperature"`
		TemperatureOutside float64 `json:"temperature_outside"`
		Date               string  `json:"date"`
		IsPublic           int     `json:"is_public"`
		IsMarkdown         int     `json:"is_markdown"`
	}
	_ = c.ShouldBindJSON(&body)
	now := util.NowString()
	diary, _ := db.Open(dbName)
	_, err := diary.Exec(`update diaries set date_modify=?, date=?, category=?, title=?, content=?, weather=?, temperature=?, temperature_outside=?, is_public=?, is_markdown=? WHERE id=? and uid=?`,
		now, body.Date, body.Category, util.UnicodeEncode(body.Title), util.UnicodeEncode(body.Content),
		body.Weather, body.Temperature, body.TemperatureOutside, body.IsPublic, body.IsMarkdown, body.ID, user.UID)
	if err != nil {
		response.Error(c, err.Error(), "日记修改失败")
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
		DiaryID int64 `json:"diaryId"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	_, err := diary.Exec(`DELETE from diaries WHERE id=? and uid=?`, body.DiaryID, user.UID)
	if err != nil {
		response.Error(c, err.Error(), "日记删除失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "删除成功")
}

func handleLatestRecommend(c *gin.Context) {
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, `select * from diaries where title like '%首页推荐%' and is_public = 1 and uid = 3 order by date desc`)
	if err != nil {
		response.Error(c, err.Error(), "查询错误")
		return
	}
	if len(rows) > 0 {
		row := rows[0]
		if t, ok := row["title"]; ok {
			row["title"] = util.UnicodeDecode(asString(t))
		}
		if ct, ok := row["content"]; ok {
			row["content"] = util.UnicodeDecode(asString(ct))
		}
		response.Success(c, row, "")
		return
	}
	response.Success(c, "", "")
}

func handleGetContentWithKeyword(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	encoded := util.UnicodeEncode(c.Query("keyword"))
	diary, _ := db.Open(dbName)
	row, err := db.QueryMap(diary, `select * from diaries where title like ? and uid = ? order by id desc`, "%"+encoded+"%", user.UID)
	if err != nil {
		response.Error(c, err.Error(), "查询错误")
		return
	}
	if row == nil {
		response.Success(c, "", "")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	decodeDiaryRow(row, true, false)
	response.Success(c, row, "")
}

func handleGetLatestPublicWithKeyword(c *gin.Context) {
	encoded := util.UnicodeEncode(c.Query("keyword"))
	diary, _ := db.Open(dbName)
	row, err := db.QueryMap(diary, `select diaries.* from diaries inner join users on diaries.uid = users.uid where diaries.title like ? and diaries.is_public = 1 and users.group_id = ? order by diaries.id desc`, "%"+encoded+"%", enumUserGroupAdmin)
	if err != nil {
		response.Error(c, err.Error(), "查询错误")
		return
	}
	if row == nil {
		response.Success(c, "", "")
		return
	}
	decodeDiaryRow(row, true, false)
	response.Success(c, row, "")
}

func handleClear(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, msg, "无权查看日记列表：用户信息错误")
		return
	}
	if systemconfig.IsConfiguredDemoAccountEmail(user.Email) {
		response.Error(c, "", "演示帐户不允许执行此操作")
		return
	}
	diary, _ := db.Open(dbName)
	res, err := diary.Exec(`delete from diaries where uid=?`, user.UID)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	affected, _ := res.RowsAffected()
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"affectedRows": affected}, fmt.Sprintf("清空成功：%d 条日记", affected))
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

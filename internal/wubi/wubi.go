package wubi

import (
	"fmt"
	"strings"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const dbWubi = db.Wubi
const dbDiary = db.Diary

// ==================== Dict (/wubi/dict) ====================

const dictTable = "wubi_dict"

func RegisterDict(r *gin.RouterGroup) {
	r.GET("/pull", dictPull)
	r.PUT("/push", dictPush)
	r.POST("/check-backup-exist", dictCheckBackupExist)
}

func dictPull(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	wubi, _ := db.Open(dbWubi)
	rows, err := db.QueryMaps(wubi, `select * from `+dictTable+` where title = ? and uid=?`, c.Query("title"), user.UID)
	if err != nil {
		response.Error(c, err.Error(), "")
		return
	}
	if len(rows) == 0 {
		response.Success(c, "", "不存在词库")
		return
	}
	row := rows[0]
	row["title"] = util.UnicodeDecode(asString(row["title"]))
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, row, "")
}

func dictPush(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body struct {
		Title       string `json:"title"`
		Content     string `json:"content"`
		ContentSize int    `json:"contentSize"`
		WordCount   int    `json:"wordCount"`
	}
	_ = c.ShouldBindJSON(&body)
	now := util.NowString()
	encodedTitle := util.UnicodeEncode(body.Title)
	wubi, _ := db.Open(dbWubi)
	exist, _ := db.QueryMaps(wubi, `select * from `+dictTable+` where title=? and uid=?`, encodedTitle, user.UID)
	if len(exist) > 0 {
		_, err := wubi.Exec(`update `+dictTable+` set title=?, content=?, content_size=?, word_count=?, date_update=? WHERE title=? and uid=?`,
			encodedTitle, body.Content, body.ContentSize, body.WordCount, now, encodedTitle, user.UID)
		if err != nil {
			response.Error(c, err.Error(), "上传失败")
			return
		}
		if diary, e := db.Open(dbDiary); e == nil {
			_, _ = diary.Exec(`update users set sync_count=sync_count + 1 WHERE uid=?`, user.UID)
		}
		util.UpdateUserLastLoginTime(user.UID)
		response.Success(c, nil, "上传成功")
		return
	}
	res, err := wubi.Exec(`INSERT into `+dictTable+`(title, content, content_size, word_count, date_init, date_update, comment, uid)
		VALUES(?,?,?,?,?,?,'',?)`,
		encodedTitle, body.Content, body.ContentSize, body.WordCount, now, now, user.UID)
	if err != nil {
		response.Error(c, err.Error(), "上传失败")
		return
	}
	id, _ := res.LastInsertId()
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": id}, "上传成功")
}

func dictCheckBackupExist(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body struct {
		FileName string `json:"fileName"`
	}
	_ = c.ShouldBindJSON(&body)
	wubi, _ := db.Open(dbWubi)
	row, err := db.QueryMap(wubi, `select id, title, content_size, word_count, date_init, date_update, comment, uid, sync_count from `+dictTable+` where title = ? and uid=?`, body.FileName, user.UID)
	if err != nil {
		response.Error(c, err.Error(), "")
		return
	}
	response.Success(c, row, "信息获取成功")
}

// ==================== Word (/wubi/word) ====================

const wordTable = "wubi_words"

func RegisterWord(r *gin.RouterGroup) {
	r.POST("/upload-dict", wordUploadDict)
	r.POST("/list", wordList)
	r.POST("/export-extra", wordExportExtra)
	r.POST("/check-exist", wordCheckExist)
	r.POST("/add", wordAdd)
	r.POST("/add-batch", wordAddBatch)
	r.PUT("/modify", wordModify)
	r.DELETE("/delete", wordDelete)
	r.PUT("/modify-batch", wordModifyBatch)
	r.GET("/statistic", func(c *gin.Context) {})
	r.GET("/thumbs-up", func(c *gin.Context) {})
	r.GET("/thumbs-down", func(c *gin.Context) {})
}

// wordUploadDict 已废弃
func wordUploadDict(c *gin.Context) {
	response.Error(c, "", "该接口已废弃")
}

func wordList(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body map[string]interface{}
	_ = c.ShouldBindJSON(&body)

	base := `
		SELECT wubi_words.id, wubi_words.word, wubi_words.code, wubi_words.priority, wubi_words.up, wubi_words.down,
		       wubi_words.date_create, wubi_words.date_modify, wubi_words.comment,
		       wubi_category.id AS category_id, wubi_category.name as category_name,
		       wubi_words.user_init, wubi_words.user_modify, wubi_words.approved,
		       users_init.uid as uid_init, users_init.nickname as nickname_init, users_init.group_id as group_id_init,
		       users_modify.uid as uid_modify, users_modify.nickname as nickname_modify, users_modify.group_id as group_id_modify
		from wubi.wubi_words
		LEFT JOIN wubi_category ON category_id = wubi_category.id
		LEFT JOIN diary.users users_init ON wubi_words.user_init = users_init.uid
		LEFT JOIN diary.users users_modify ON wubi_words.user_modify = users_modify.uid`

	var filters []string
	var args []interface{}

	if kw := toStr(body["keyword"]); kw != "" {
		var parts []string
		for _, k := range strings.Split(kw, " ") {
			parts = append(parts, "( wubi_words.word like ? ESCAPE '/'  or  wubi_words.code like ? ESCAPE '/' or wubi_words.comment like ? ESCAPE '/')")
			like := "%" + util.UnicodeEncode(k) + "%"
			args = append(args, like, like, like)
		}
		filters = append(filters, strings.Join(parts, " and "))
	}
	if dr, ok := body["dateRange"].([]interface{}); ok && len(dr) == 2 {
		filters = append(filters, "date_create between ? AND ?")
		args = append(args, toStr(dr[0]), toStr(dr[1]))
	}
	if truthy(body["category_id"]) {
		filters = append(filters, "category_id = ?")
		args = append(args, toInt(body["category_id"]))
	}
	if v, ok := body["approved"]; ok && toStr(v) != "" {
		filters = append(filters, "wubi_words.approved = ?")
		args = append(args, toInt(v))
	}

	whereSQL := ""
	if len(filters) > 0 {
		whereSQL = "where " + strings.Join(filters, " and ")
	}
	orderSQL := " order by wubi_words.code asc, wubi_words.priority asc"

	pageNo := toInt(body["pageNo"])
	pageSize := toInt(body["pageSize"])
	if pageNo < 1 {
		pageNo = 1
	}
	pointStart := (pageNo - 1) * pageSize

	wubi, _ := db.Open(dbWubi)
	listArgs := append(append([]interface{}{}, args...), pointStart, pageSize)
	list, err := db.QueryMaps(wubi, base+" "+whereSQL+orderSQL+" limit ? , ?", listArgs...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	countRow, err := db.QueryMap(wubi, "select count(*) as sum from "+wordTable+" "+whereSQL, args...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	for _, item := range list {
		item["word"] = util.UnicodeDecode(asString(item["word"]))
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

func wordExportExtra(c *gin.Context) {
	_, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	wubi, _ := db.Open(dbWubi)
	list, err := db.QueryMaps(wubi, `
		SELECT wubi_words.id, wubi_words.word, wubi_words.code, wubi_words.priority, wubi_words.comment,
		       wubi_category.id AS category_id, wubi_category.name as category_name,
		       wubi_words.user_init, wubi_words.user_modify, diary.users.email, diary.users.group_id
		FROM wubi_words
		LEFT JOIN wubi_category ON category_id = wubi_category.id
		LEFT JOIN diary.users ON wubi_words.user_init = diary.users.uid
		WHERE category_id != 1 and approved = 1
		ORDER BY wubi_category.sort_id, wubi_words.id ASC`)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	for _, item := range list {
		item["word"] = util.UnicodeDecode(asString(item["word"]))
	}
	response.Success(c, list, "请求成功")
}

func wordCheckExist(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		Word string `json:"word"`
		Code string `json:"code"`
	}
	_ = c.ShouldBindJSON(&body)
	wubi, _ := db.Open(dbWubi)
	list, err := db.QueryMaps(wubi, `select * from `+wordTable+` where word like ? and code like ? limit 5`,
		"%"+util.UnicodeEncode(body.Word)+"%", body.Code+"%")
	if err != nil {
		response.Error(c, err.Error(), "查询失败")
		return
	}
	response.Success(c, list, "查询成功")
	util.UpdateUserLastLoginTime(user.UID)
}

func wordAdd(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		Word       string `json:"word"`
		Code       string `json:"code"`
		Priority   int    `json:"priority"`
		Up         int    `json:"up"`
		Down       int    `json:"down"`
		Comment    string `json:"comment"`
		CategoryID int    `json:"category_id"`
	}
	_ = c.ShouldBindJSON(&body)
	now := util.NowString()
	isApproved := 0
	if user.IsAdmin() {
		isApproved = 1
	}
	categoryID := body.CategoryID
	if categoryID == 0 {
		categoryID = 1
	}
	wubi, _ := db.Open(dbWubi)
	res, err := wubi.Exec(`INSERT into `+wordTable+`(word, code, priority, up, down, date_create, date_modify, comment, user_init, user_modify, category_id, approved)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		util.UnicodeEncode(body.Word), body.Code, body.Priority, body.Up, body.Down, now, now, body.Comment, user.UID, user.UID, categoryID, isApproved)
	if err != nil {
		response.Error(c, err.Error(), "添加失败")
		return
	}
	id, _ := res.LastInsertId()
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": id}, "添加成功")
}

func wordAddBatch(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		Words []struct {
			Word     string `json:"word"`
			Code     string `json:"code"`
			Priority int    `json:"priority"`
			Comment  string `json:"comment"`
		} `json:"words"`
		CategoryID int `json:"category_id"`
	}
	_ = c.ShouldBindJSON(&body)
	wubi, _ := db.Open(dbWubi)

	categoryID := body.CategoryID
	if categoryID == 0 {
		categoryID = 1
	}
	isApproved := 0
	if user.IsAdmin() {
		isApproved = 1
	}
	now := util.NowString()

	existCount := 0
	addedCount := 0
	for _, w := range body.Words {
		existRows, _ := db.QueryMaps(wubi, `select * from `+wordTable+` where code = ? and word = ?`, w.Code, w.Word)
		if len(existRows) > 0 {
			existCount++
			continue
		}
		_, err := wubi.Exec(`INSERT into `+wordTable+`(word, code, priority, date_create, date_modify, comment, user_init, user_modify, category_id, approved)
			VALUES(?,?,?,?,?,?,?,?,?,?)`,
			util.UnicodeEncode(w.Word), w.Code, w.Priority, now, now, w.Comment, user.UID, user.UID, categoryID, isApproved)
		if err != nil {
			response.Error(c, err.Error(), "添加失败")
			return
		}
		addedCount++
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"addedCount": addedCount, "existCount": existCount}, "批量添加成功")
}

func wordModify(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		ID         int64  `json:"id"`
		Word       string `json:"word"`
		Code       string `json:"code"`
		Priority   int    `json:"priority"`
		Up         int    `json:"up"`
		Down       int    `json:"down"`
		Comment    string `json:"comment"`
		CategoryID int    `json:"category_id"`
	}
	_ = c.ShouldBindJSON(&body)
	now := util.NowString()
	isApproved := 0
	if user.IsAdmin() {
		isApproved = 1
	}
	wubi, _ := db.Open(dbWubi)
	query := `update ` + wordTable + ` set date_modify=?, word=?, code=?, priority=?, up=?, down=?, comment=?, category_id=?, user_modify=?, approved=? WHERE id=?`
	args := []interface{}{now, util.UnicodeEncode(body.Word), body.Code, body.Priority, body.Up, body.Down, body.Comment, body.CategoryID, user.UID, isApproved, body.ID}
	if !user.IsAdmin() {
		query += ` and user_init = ?`
		args = append(args, user.UID)
	}
	_, err := wubi.Exec(query, args...)
	if err != nil {
		response.Error(c, err.Error(), "修改失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "修改成功")
}

func wordDelete(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		IDs []int64 `json:"ids"`
	}
	_ = c.ShouldBindJSON(&body)
	if len(body.IDs) == 0 {
		response.Error(c, "", "参数错误：ids 为空")
		return
	}
	placeholders := make([]string, len(body.IDs))
	args := make([]interface{}, 0, len(body.IDs)+1)
	for i, id := range body.IDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	query := `DELETE from ` + wordTable + ` WHERE id in (` + strings.Join(placeholders, ",") + `)`
	if !user.IsAdmin() {
		query += ` and user_init = ?`
		args = append(args, user.UID)
	}
	wubi, _ := db.Open(dbWubi)
	_, err := wubi.Exec(query, args...)
	if err != nil {
		response.Error(c, err.Error(), "五笔词条删除失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "删除成功")
}

func wordModifyBatch(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	var body map[string]interface{}
	_ = c.ShouldBindJSON(&body)
	now := util.NowString()

	setParts := []string{"date_modify=?", "user_modify=?"}
	args := []interface{}{now, user.UID}
	if _, ok := body["category_id"]; ok {
		setParts = append(setParts, "category_id=?")
		args = append(args, toInt(body["category_id"]))
	}
	if _, ok := body["approved"]; ok {
		setParts = append(setParts, "approved=?")
		args = append(args, toInt(body["approved"]))
	}

	ids, _ := body["ids"].([]interface{})
	if len(ids) == 0 {
		response.Error(c, "", "参数错误：ids 为空")
		return
	}
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, toInt(id))
	}
	query := `update ` + wordTable + ` set ` + strings.Join(setParts, ", ") + ` WHERE id in (` + strings.Join(placeholders, ",") + `)`
	wubi, _ := db.Open(dbWubi)
	_, err := wubi.Exec(query, args...)
	if err != nil {
		response.Error(c, err.Error(), "修改失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "修改成功")
}

// ==================== Category (/wubi/category) ====================

const categoryTable = "wubi_category"

func RegisterCategory(r *gin.RouterGroup) {
	r.GET("/list", categoryList)
	r.POST("/add", categoryAdd)
	r.PUT("/modify", categoryModify)
	r.DELETE("/delete", categoryDelete)
}

func categoryList(c *gin.Context) {
	_, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	wubi, _ := db.Open(dbWubi)
	categories, err := db.QueryMaps(wubi, `select * from `+categoryTable+` order by sort_id asc`)
	if err != nil {
		response.Error(c, err.Error(), "")
		return
	}
	// 统计每个类别下的词条数量
	var countParts []string
	for _, cat := range categories {
		id := toInt(cat["id"])
		countParts = append(countParts, fmt.Sprintf("count(case when category_id=%d then 1 end) as '%d'", id, id))
	}
	if len(countParts) > 0 {
		countRow, err := db.QueryMap(wubi, `select `+strings.Join(countParts, ", ")+`, count(*) as amount from wubi_words`)
		if err == nil && countRow != nil {
			for _, cat := range categories {
				id := toInt(cat["id"])
				cat["count"] = countRow[fmt.Sprintf("%d", id)]
			}
		}
	}
	response.Success(c, categories, "")
}

func categoryAdd(c *gin.Context) {
	var body struct {
		Name   string `json:"name"`
		SortID int    `json:"sort_id"`
	}
	_ = c.ShouldBindJSON(&body)
	wubi, _ := db.Open(dbWubi)
	exist, _ := db.QueryMaps(wubi, `select * from `+categoryTable+` where name=?`, body.Name)
	if len(exist) > 0 {
		response.Error(c, "", "五笔码表类别已存在")
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
	res, err := wubi.Exec(`insert into `+categoryTable+`(name, sort_id, date_init) values (?, ?, ?)`, body.Name, body.SortID, now)
	if err != nil {
		response.Error(c, err.Error(), "五笔码表类别添加失败")
		return
	}
	id, _ := res.LastInsertId()
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": id}, "添加成功")
}

func categoryModify(c *gin.Context) {
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
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		SortID int    `json:"sort_id"`
	}
	_ = c.ShouldBindJSON(&body)
	wubi, _ := db.Open(dbWubi)
	res, err := wubi.Exec(`update `+categoryTable+` set name = ?, sort_id = ? where id = ?`, body.Name, body.SortID, body.ID)
	if err != nil {
		response.Error(c, err.Error(), "五笔码表类别修改失败")
		return
	}
	id, _ := res.LastInsertId()
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": id}, "修改成功")
}

func categoryDelete(c *gin.Context) {
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
		ID int64 `json:"id"`
	}
	_ = c.ShouldBindJSON(&body)
	wubi, _ := db.Open(dbWubi)
	_, err := wubi.Exec(`delete from `+categoryTable+` where id = ?`, body.ID)
	if err != nil {
		response.Error(c, err.Error(), "五笔码表类别删除失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "删除成功")
}

// ==================== helpers ====================

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

func toStr(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case float64:
		return fmt.Sprintf("%v", s)
	default:
		return fmt.Sprintf("%v", s)
	}
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case string:
		var x int
		fmt.Sscan(n, &x)
		return x
	}
	return 0
}

func truthy(v interface{}) bool {
	switch n := v.(type) {
	case nil:
		return false
	case float64:
		return n != 0
	case int:
		return n != 0
	case string:
		return n != "" && n != "0"
	case bool:
		return n
	}
	return true
}

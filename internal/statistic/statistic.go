package statistic

import (
	"fmt"
	"strings"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/gin-gonic/gin"
)

const dbName = db.Diary

func Register(r *gin.RouterGroup) {
	r.GET("/", handleOverview)
	r.GET("/user-data-diary", handleUserDataDiary)
	r.GET("/user-data-words", handleUserDataWords)
	r.GET("/category", handleCategory)
	r.GET("/year", handleYear)
	// /users 已下线（原 diary 端用户统计）；改用管理员专用 /manager-users
	r.GET("/manager-users", handleManagerUsers)
	r.GET("/weather", handleWeather)
}

func handleOverview(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	diary, _ := db.Open(dbName)
	var row map[string]interface{}
	var err error
	if user.IsAdmin() {
		row, err = db.QueryMap(diary, `SELECT
			(SELECT COUNT(*) FROM diaries) as count_diary,
			(SELECT COUNT(*) FROM qrs) as count_qr,
			(SELECT COUNT(*) FROM users) as count_user,
			(SELECT COUNT(*) FROM diary_category) as count_category,
			(SELECT COUNT(*) FROM diaries where category = 'bill') as count_bill,
			(SELECT COUNT(*) FROM wubi.wubi_dict) as count_dict,
			(SELECT COUNT(*) FROM wubi.wubi_words ) as count_wubi_words,
			(SELECT COUNT(*) FROM wubi.wubi_words where approved = 0) as count_wubi_words_unapproved,
			(SELECT COUNT(*) FROM wubi.wubi_words where approved = 0 and user_init = ? ) as count_wubi_words_unapproved_user`, user.UID)
	} else {
		row, err = db.QueryMap(diary, `SELECT
			(SELECT COUNT(*) FROM diaries where uid = ?) as count_diary,
			(SELECT COUNT(*) FROM qrs where uid = ?) as count_qr,
			(SELECT COUNT(*) FROM users where uid = ?) as count_user,
			(SELECT COUNT(*) FROM diary_category) as count_category,
			(SELECT COUNT(*) FROM diaries where uid = ? and category = 'bill') as count_bill,
			(SELECT COUNT(*) FROM wubi.wubi_dict where uid = ?) as count_dict,
			(SELECT COUNT(*) FROM wubi.wubi_words ) as count_wubi_words,
			(SELECT COUNT(*) FROM wubi.wubi_words where approved = 0) as count_wubi_words_unapproved,
			(SELECT COUNT(*) FROM wubi.wubi_words where approved = 0 and user_init = ? ) as count_wubi_words_unapproved_user`,
			user.UID, user.UID, user.UID, user.UID, user.UID, user.UID)
	}
	if err != nil {
		response.Error(c, "", err.Error())
		return
	}
	response.Success(c, row, "")
}

func handleUserDataDiary(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "需要管理员权限")
		return
	}
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, `SELECT nickname, count_diary FROM users where count_diary > 3 order by count_diary desc`)
	if err != nil {
		response.Error(c, "", err.Error())
		return
	}
	response.Success(c, rows, "")
}

func handleUserDataWords(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "需要管理员权限")
		return
	}
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, `SELECT nickname, count_words FROM users where count_words != 0 order by count_words desc`)
	if err != nil {
		response.Error(c, "", err.Error())
		return
	}
	response.Success(c, rows, "")
}

func handleCategory(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	diary, _ := db.Open(dbName)
	categories, err := db.QueryMaps(diary, `select * from diary_category order by sort_id asc`)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	var parts []string
	for _, cat := range categories {
		nameEn := asString(cat["name_en"])
		parts = append(parts, fmt.Sprintf("count(case when category='%s' then 1 end) as %s", nameEn, nameEn))
	}
	query := `select ` + strings.Join(parts, ", ") + `, count(case when is_public='1' then 1 end) as shared, count(*) as amount from diaries where uid=?`
	row, err := db.QueryMap(diary, query, user.UID)
	if err != nil {
		response.Error(c, err.Error(), "")
		return
	}
	response.Success(c, row, "")
}

func handleYear(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, `
		select year(date) as year, date_format(date, '%m') as month, count(*) as count
		from diaries where uid = ?
		group by year(date), date_format(date, '%m')
		order by year(date) asc, month desc`, user.UID)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}

	type monthItem struct {
		Month string `json:"month"`
		Count int64  `json:"count"`
		ID    string `json:"id"`
	}
	type yearItem struct {
		Year   int64       `json:"year"`
		Count  int64       `json:"count"`
		Months []monthItem `json:"months"`
	}
	order := []int64{}
	yearMap := map[int64]*yearItem{}
	for _, item := range rows {
		year := asInt(item["year"])
		if _, ok := yearMap[year]; !ok {
			yearMap[year] = &yearItem{Year: year, Months: []monthItem{}}
			order = append(order, year)
		}
		count := asInt(item["count"])
		month := asString(item["month"])
		yi := yearMap[year]
		yi.Count += count
		yi.Months = append(yi.Months, monthItem{Month: month, Count: count, ID: fmt.Sprintf("%d%s", year, month)})
	}
	result := make([]*yearItem, 0, len(order))
	for _, y := range order {
		result = append(result, yearMap[y])
	}
	response.Success(c, result, "")
}

// handleManagerUsers returns active-ish users for the manager statistics page.
// Admin only. Replaces the old diary-facing GET /statistic/users.
func handleManagerUsers(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "需要管理员权限")
		return
	}
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, `
		SELECT uid, last_visit_time, nickname, register_time,
			count_diary, count_dict, count_map_route, count_words, sync_count
		FROM users
		WHERE count_diary >= 1 OR count_dict >= 1 OR count_map_route >= 1 OR sync_count >= 1
		ORDER BY count_diary DESC, sync_count DESC`)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}

	diaryUsers := make([]map[string]interface{}, 0)
	dictUsers := make([]map[string]interface{}, 0)
	mapRouteUsers := make([]map[string]interface{}, 0)
	diaryChart := make([]gin.H, 0)
	for _, row := range rows {
		if asInt(row["count_diary"]) > 0 {
			diaryUsers = append(diaryUsers, row)
			diaryChart = append(diaryChart, gin.H{
				"name":  asString(row["nickname"]),
				"value": asInt(row["count_diary"]),
			})
		}
		if asInt(row["count_dict"]) > 0 {
			dictUsers = append(dictUsers, row)
		}
		if asInt(row["count_map_route"]) > 0 {
			mapRouteUsers = append(mapRouteUsers, row)
		}
	}

	response.Success(c, gin.H{
		"users":            rows,
		"diary_users":      diaryUsers,
		"dict_users":       dictUsers,
		"map_route_users":  mapRouteUsers,
		"diary_chart":      diaryChart,
	}, "请求成功")
}

func handleWeather(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, `select temperature, temperature_outside, date from diaries where category = 'life' and uid = ?`, user.UID)
	if err != nil {
		response.Error(c, err.Error(), "数据库请求错误")
		return
	}
	response.Success(c, rows, "请求成功")
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

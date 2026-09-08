package starve

import (
	"fmt"
	"strings"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const dbName = db.Starve

type listConfig struct {
	path         string
	table        string
	searchFields []string
	hasVersion   bool
}

type entityWrite struct {
	seg      string
	table    string
	cols     []string // insert/update columns (excluding id)
	required []string // required on add
	reqMsg   string
}

var listConfigs = []listConfig{
	{"/character/list", "characters", []string{"name", "name_en"}, true},
	{"/mob/list", "mobs", []string{"name", "name_en"}, true},
	{"/log/list", "logs", []string{"detail"}, false},
	{"/plant/list", "plants", []string{"name", "name_en"}, true},
	{"/thing/list", "things", []string{"name", "name_en"}, true},
	{"/material/list", "materials", []string{"name", "name_en"}, true},
	{"/craft/list", "crafts", []string{"name", "name_en"}, true},
	{"/cookingrecipe/list", "cookrecipes", []string{"name", "name_en"}, true},
	{"/coder/list", "coders", []string{"usage", "code", "note"}, false},
}

var infoConfigs = []struct{ path, table string }{
	{"/character/info", "characters"},
	{"/mob/info", "mobs"},
	{"/log/info", "logs"},
	{"/plant/info", "plants"},
	{"/thing/info", "things"},
	{"/material/info", "materials"},
	{"/craft/info", "crafts"},
	{"/cookreceipe/info", "cookrecipes"},
	{"/coder/info", "coders"},
}

var writeEntities = []entityWrite{
	{
		seg: "log", table: "logs",
		cols: []string{"date", "detail"},
		required: []string{"date", "detail"}, reqMsg: "日期和详情不能为空",
	},
	{
		seg: "character", table: "characters",
		cols: []string{
			"name", "name_en", "nick_name", "motto", "perk", "health", "hunger", "sanity",
			"hunger_modifier", "sanity_modifier", "wetness_modifier",
			"health_range", "damage_range", "hunger_range", "sanity_range", "speed_range",
			"debugspawn", "pic", "thumb", "special_item", "starting_item", "version",
		},
		required: []string{"health", "hunger", "sanity", "version"},
		reqMsg:   "生命值、饥饿值、理智值和版本不能为空",
	},
	{
		seg: "coder", table: "coders",
		cols: []string{"usage", "code", "note"},
		required: []string{"usage", "code"}, reqMsg: "用途和代码不能为空",
	},
	{
		seg: "cookingrecipe", table: "cookrecipes",
		cols: []string{
			"name", "name_en", "hunger", "health", "sanity", "spoil", "cooktime",
			"priority", "foodtype", "note", "pic", "thumb", "version", "is_active",
		},
		required: []string{"name"}, reqMsg: "名称不能为空",
	},
	{
		seg: "craft", table: "crafts",
		cols: []string{
			"name", "name_en", "tab", "tier", "ingredient", "description",
			"pic", "thumb", "version", "is_active",
		},
		required: []string{"name"}, reqMsg: "名称不能为空",
	},
	{
		seg: "material", table: "materials",
		cols: []string{"name", "name_en", "type", "pic", "thumb", "version", "is_active"},
		required: []string{"name"}, reqMsg: "名称不能为空",
	},
	{
		seg: "mob", table: "mobs",
		cols: []string{
			"name", "name_en", "health", "damage", "attack_period", "attack_range",
			"run_speed", "walk_speed", "sanity_aura", "kind", "size",
			"pic", "thumb", "version", "is_active",
		},
		required: []string{"name"}, reqMsg: "名称不能为空",
	},
	{
		seg: "plant", table: "plants",
		cols: []string{"name", "name_en", "pic", "thumb", "version", "is_active"},
		required: []string{"name"}, reqMsg: "名称不能为空",
	},
	{
		seg: "thing", table: "things",
		cols: []string{"name", "name_en", "pic", "thumb", "version", "is_active"},
		required: []string{"name"}, reqMsg: "名称不能为空",
	},
}

func Register(r *gin.RouterGroup) {
	for _, cfg := range listConfigs {
		cfg := cfg
		r.GET(cfg.path, func(c *gin.Context) { handleList(c, cfg) })
	}
	for _, cfg := range infoConfigs {
		cfg := cfg
		r.GET(cfg.path, func(c *gin.Context) { handleInfo(c, cfg.table) })
	}
	for _, e := range writeEntities {
		e := e
		r.POST("/"+e.seg+"/add", func(c *gin.Context) { handleAdd(c, e) })
		r.PUT("/"+e.seg+"/modify", func(c *gin.Context) { handleModify(c, e) })
		r.DELETE("/"+e.seg+"/delete", func(c *gin.Context) { handleDelete(c, e.table) })
	}
}

func handleList(c *gin.Context, cfg listConfig) {
	var conditions []string
	var args []interface{}

	if kw := c.Query("keyword"); kw != "" && len(cfg.searchFields) > 0 {
		encoded := util.UnicodeEncode(kw)
		var parts []string
		for _, field := range cfg.searchFields {
			parts = append(parts, field+" LIKE ? ESCAPE '/'")
			args = append(args, "%"+encoded+"%")
		}
		conditions = append(conditions, "("+strings.Join(parts, " OR ")+")")
	}
	if cfg.hasVersion {
		if v := c.Query("version"); v != "" {
			conditions = append(conditions, "version = ?")
			args = append(args, v)
		}
	}
	// Classic starve schema on this host has no is_active on most tables;
	// only apply the filter when explicitly requested via ?active=1.
	if c.Query("active") == "1" {
		conditions = append(conditions, "is_active = 1")
	}
	// Preserve Node behavior when caller passes all=0 and schema supports it:
	if c.Query("all") == "0" {
		conditions = append(conditions, "is_active = 1")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	starveDB, err := db.Open(dbName)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}

	pageNoStr := c.Query("pageNo")
	pageSizeStr := c.Query("pageSize")
	if pageNoStr != "" && pageSizeStr != "" {
		pageNo := util.AtoiDefault(pageNoStr, 1)
		pageSize := util.AtoiDefault(pageSizeStr, 20)
		startPoint := (pageNo - 1) * pageSize
		listArgs := append(append([]interface{}{}, args...), startPoint, pageSize)
		list, err := db.QueryMaps(starveDB, `SELECT * FROM `+cfg.table+` `+whereClause+` LIMIT ?, ?`, listArgs...)
		if err != nil {
			response.Error(c, err.Error(), err.Error())
			return
		}
		countRow, _ := db.QueryMap(starveDB, `SELECT COUNT(*) as sum FROM `+cfg.table+` `+whereClause, args...)
		var total interface{} = 0
		if countRow != nil {
			total = countRow["sum"]
		}
		response.Success(c, gin.H{
			"list":  list,
			"pager": gin.H{"pageSize": pageSize, "pageNo": pageNo, "total": total},
		}, "请求成功")
		return
	}

	list, err := db.QueryMaps(starveDB, `SELECT * FROM `+cfg.table+` `+whereClause, args...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	response.Success(c, list, "")
}

func handleInfo(c *gin.Context, table string) {
	starveDB, err := db.Open(dbName)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	row, err := db.QueryMap(starveDB, `SELECT * FROM `+table+` WHERE id = ?`, c.Query("id"))
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	if row == nil {
		response.Error(c, "", "无数据")
		return
	}
	response.Success(c, row, "")
}

func handleAdd(c *gin.Context, e entityWrite) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, err.Error(), "参数错误")
		return
	}
	for _, req := range e.required {
		if empty(body[req]) {
			response.Error(c, "", e.reqMsg)
			return
		}
	}
	cols := []string{}
	ph := []string{}
	args := []interface{}{}
	for _, col := range e.cols {
		if v, ok := body[col]; ok {
			cols = append(cols, col)
			ph = append(ph, "?")
			args = append(args, encodeField(v))
		}
	}
	if len(cols) == 0 {
		response.Error(c, "", "无有效字段")
		return
	}
	starveDB, err := db.Open(dbName)
	if err != nil {
		response.Error(c, err.Error(), "添加失败")
		return
	}
	q := fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s)", e.table, strings.Join(cols, ","), strings.Join(ph, ","))
	res, err := starveDB.Exec(q, args...)
	if err != nil {
		response.Error(c, err.Error(), "添加失败")
		return
	}
	id, _ := res.LastInsertId()
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": id}, "添加成功")
}

func handleModify(c *gin.Context, e entityWrite) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, err.Error(), "参数错误")
		return
	}
	if empty(body["id"]) {
		response.Error(c, "", "ID不能为空")
		return
	}
	sets := []string{}
	args := []interface{}{}
	for _, col := range e.cols {
		if v, ok := body[col]; ok {
			sets = append(sets, col+"=?")
			args = append(args, encodeField(v))
		}
	}
	if len(sets) == 0 {
		response.Error(c, "", "无有效字段")
		return
	}
	args = append(args, body["id"])
	starveDB, err := db.Open(dbName)
	if err != nil {
		response.Error(c, err.Error(), "修改失败")
		return
	}
	q := fmt.Sprintf("UPDATE %s SET %s WHERE id=?", e.table, strings.Join(sets, ","))
	if _, err := starveDB.Exec(q, args...); err != nil {
		response.Error(c, err.Error(), "修改失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "修改成功")
}

func handleDelete(c *gin.Context, table string) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body struct {
		ID interface{} `json:"id"`
	}
	_ = c.ShouldBindJSON(&body)
	id := body.ID
	if id == nil || id == "" {
		id = c.Query("id")
	}
	if empty(id) {
		response.Error(c, "", "ID不能为空")
		return
	}
	starveDB, err := db.Open(dbName)
	if err != nil {
		response.Error(c, err.Error(), "删除失败")
		return
	}
	if _, err := starveDB.Exec(`DELETE FROM `+table+` WHERE id=?`, id); err != nil {
		response.Error(c, err.Error(), "删除失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "删除成功")
}

func empty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t) == ""
	case float64:
		return false
	default:
		return fmt.Sprint(v) == ""
	}
}

func encodeField(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		if strings.TrimSpace(t) == "" {
			return nil
		}
		return util.UnicodeEncode(t)
	case float64, int, int64, bool:
		return t
	default:
		s := fmt.Sprint(t)
		if strings.TrimSpace(s) == "" {
			return nil
		}
		return util.UnicodeEncode(s)
	}
}

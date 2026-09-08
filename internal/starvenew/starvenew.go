// Package starvenew ports src/dontstarve/dontStarveNew.ts to Go. It serves the
// /starve-new route group backed by the `starve_advance` database. Compared to
// the classic starve module, versions and craft tabs are modelled as separate
// entities linked through junction tables (character_versions, craft_versions,
// craft_tab_relations, ...). GET list/info endpoints are public; mutations
// require authorization via middleware.VerifyAuthorization.
package starvenew

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/KyleBing/portal-go/internal/apihelper"
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const dbName = db.StarveAdvance

type colKind int

const (
	kEnc colKind = iota // unicode-encoded, NULL when empty
	kRaw                // raw string, always present
	kNum                // number, NULL when empty
)

type col struct {
	name string
	kind colKind
}

// junction describes a many-to-many relation persisted through a link table.
type junction struct {
	field     string // request-body field: "version" or "tab"
	table     string // link table: character_versions / craft_tab_relations
	entityCol string // FK to the owning entity: character_id / craft_id
	refCol    string // FK to the referenced entity: version_id / tab_id
}

type entity struct {
	seg             string
	table           string
	cols            []col
	hasIsActive     bool
	isActiveDefault int
	junctions       []junction
	reqAdd          []string // fields required by /add ("version"/"tab" checked as id lists)
	reqModify       []string // extra fields required by /modify (beyond id)
	reqMsg          string
}

type listConfig struct {
	path       string
	table      string
	search     []string
	hasVersion bool
	hasTab     bool
	orderBy    string // optional ORDER BY (used by version / craft-tab lists)
}

type infoConfig struct {
	path       string
	table      string
	hasVersion bool
	hasTab     bool
}

// Register wires all /starve-new routes onto the provided router group.
func Register(r *gin.RouterGroup) {
	for _, lc := range listConfigs {
		r.GET(lc.path, listHandler(lc))
	}
	for _, ic := range infoConfigs {
		r.GET(ic.path, infoHandler(ic))
	}
	for _, e := range entities {
		e := e
		r.POST("/"+e.seg+"/add", addHandler(e))
		r.PUT("/"+e.seg+"/modify", modifyHandler(e))
		r.DELETE("/"+e.seg+"/delete", deleteHandler(e))
	}
}

// ---------------------------------------------------------------------------
// configuration
// ---------------------------------------------------------------------------

var listConfigs = []listConfig{
	{path: "/character/list", table: "characters", search: []string{"name", "name_en"}, hasVersion: true},
	{path: "/mob/list", table: "mobs", search: []string{"name", "name_en"}, hasVersion: true},
	{path: "/log/list", table: "logs", search: []string{"detail"}},
	{path: "/plant/list", table: "plants", search: []string{"name", "name_en"}, hasVersion: true},
	{path: "/thing/list", table: "things", search: []string{"name", "name_en"}, hasVersion: true},
	{path: "/material/list", table: "materials", search: []string{"name", "name_en"}},
	{path: "/craft/list", table: "crafts", search: []string{"name", "name_en"}, hasVersion: true, hasTab: true},
	{path: "/cookingrecipe/list", table: "cookingrecipes", search: []string{"name", "name_en"}},
	{path: "/coder/list", table: "coders", search: []string{"usage", "code", "note"}},
	{path: "/version/list", table: "versions", search: []string{"code", "name", "name_en"}, orderBy: "sort_order ASC, id ASC"},
	{path: "/craft-tab/list", table: "craft_tabs", search: []string{"code", "name", "name_en"}, orderBy: "sort_order ASC, id ASC"},
}

var infoConfigs = []infoConfig{
	{path: "/character/info", table: "characters", hasVersion: true},
	{path: "/mob/info", table: "mobs", hasVersion: true},
	{path: "/log/info", table: "logs"},
	{path: "/plant/info", table: "plants", hasVersion: true},
	{path: "/thing/info", table: "things", hasVersion: true},
	{path: "/material/info", table: "materials"},
	{path: "/craft/info", table: "crafts", hasVersion: true, hasTab: true},
	{path: "/cookingrecipe/info", table: "cookingrecipes"},
	{path: "/coder/info", table: "coders"},
	{path: "/version/info", table: "versions"},
	{path: "/craft-tab/info", table: "craft_tabs"},
}

var (
	versionJunctionCharacter = junction{"version", "character_versions", "character_id", "version_id"}
	versionJunctionMob       = junction{"version", "mob_versions", "mob_id", "version_id"}
	versionJunctionPlant     = junction{"version", "plant_versions", "plant_id", "version_id"}
	versionJunctionThing     = junction{"version", "thing_versions", "thing_id", "version_id"}
	versionJunctionCraft     = junction{"version", "craft_versions", "craft_id", "version_id"}
	tabJunctionCraft         = junction{"tab", "craft_tab_relations", "craft_id", "tab_id"}
)

var entities = []entity{
	{
		seg: "log", table: "logs",
		cols:      []col{{"date", kRaw}, {"detail", kEnc}},
		reqAdd:    []string{"date", "detail"},
		reqModify: []string{"date", "detail"},
		reqMsg:    "日期和详情不能为空",
	},
	{
		seg: "character", table: "characters",
		cols: []col{
			{"name", kEnc}, {"name_en", kEnc}, {"nick_name", kEnc}, {"motto", kEnc}, {"perk", kEnc},
			{"health", kRaw}, {"hunger", kRaw}, {"sanity", kRaw},
			{"hunger_modifier", kEnc}, {"sanity_modifier", kEnc}, {"wetness_modifier", kEnc},
			{"health_range", kEnc}, {"damage_range", kEnc}, {"hunger_range", kEnc}, {"sanity_range", kEnc}, {"speed_range", kEnc},
			{"debugspawn", kEnc}, {"pic", kEnc}, {"thumb", kEnc}, {"special_item", kEnc}, {"starting_item", kEnc},
		},
		hasIsActive: true, isActiveDefault: 0,
		junctions: []junction{versionJunctionCharacter},
		reqAdd:    []string{"health", "hunger", "sanity", "version"},
		reqMsg:    "生命值、饥饿值、理智值和版本不能为空",
	},
	{
		seg: "coder", table: "coders",
		cols:        []col{{"usage", kEnc}, {"code", kEnc}, {"note", kEnc}},
		hasIsActive: true, isActiveDefault: 0,
		reqAdd: []string{"usage", "code"},
		reqMsg: "用途和代码不能为空",
	},
	{
		seg: "cookingrecipe", table: "cookingrecipes",
		cols: []col{
			{"name", kEnc}, {"name_en", kEnc},
			{"health_value", kNum}, {"hungry_value", kNum}, {"sanity_value", kNum}, {"duration", kNum},
			{"cook_time", kEnc}, {"priority", kEnc}, {"requirements", kEnc}, {"restrictions", kEnc},
			{"perk", kEnc}, {"stacks", kEnc}, {"debugspawn", kEnc}, {"pic", kEnc}, {"thumb", kEnc},
		},
		hasIsActive: true, isActiveDefault: 0,
		reqAdd: []string{"name", "name_en"},
		reqMsg: "名称和英文名称不能为空",
	},
	{
		seg: "craft", table: "crafts",
		cols: []col{
			{"name", kEnc}, {"name_en", kEnc}, {"sortid", kNum},
			{"crafting", kEnc}, {"tier", kEnc}, {"damage", kEnc}, {"sideeffect", kEnc}, {"durability", kEnc},
			{"perk", kEnc}, {"stacks", kEnc}, {"debugspawn", kEnc}, {"pic", kEnc}, {"thumb", kEnc},
		},
		hasIsActive: true, isActiveDefault: 0,
		junctions: []junction{versionJunctionCraft, tabJunctionCraft},
		reqAdd:    []string{"name", "name_en", "crafting"},
		reqMsg:    "名称、英文名称和制作材料不能为空",
	},
	{
		seg: "material", table: "materials",
		cols:        []col{{"name", kEnc}, {"name_en", kEnc}, {"pic", kEnc}, {"stack", kEnc}, {"debugspawn", kEnc}},
		hasIsActive: true, isActiveDefault: 0,
		reqAdd: []string{"name", "name_en"},
		reqMsg: "名称和英文名称不能为空",
	},
	{
		seg: "mob", table: "mobs",
		cols: []col{
			{"name", kEnc}, {"name_en", kEnc}, {"health", kEnc}, {"damage", kEnc},
			{"attack_period", kEnc}, {"attack_range", kEnc}, {"walking_speed", kEnc}, {"running_speed", kEnc},
			{"sanityaura", kEnc}, {"special_ability", kEnc}, {"detail", kEnc}, {"loot", kEnc}, {"spawns_from", kEnc},
			{"debugspawn", kEnc}, {"pic", kEnc}, {"thumb", kEnc}, {"kind", kRaw}, {"size", kRaw},
		},
		hasIsActive: true, isActiveDefault: 0,
		junctions: []junction{versionJunctionMob},
		reqAdd:    []string{"name", "name_en", "kind", "size"},
		reqMsg:    "名称、英文名称、类型和大小不能为空",
	},
	{
		seg: "plant", table: "plants",
		cols: []col{
			{"name", kEnc}, {"name_en", kEnc}, {"resources", kEnc}, {"spawns", kEnc}, {"debugspawn", kEnc},
			{"perk", kEnc}, {"pic", kEnc}, {"thumb", kEnc},
		},
		hasIsActive: true, isActiveDefault: 0,
		junctions: []junction{versionJunctionPlant},
		reqAdd:    []string{"name", "name_en"},
		reqMsg:    "名称和英文名称不能为空",
	},
	{
		seg: "thing", table: "things",
		cols: []col{
			{"name", kEnc}, {"name_en", kEnc}, {"note", kEnc}, {"debugspawn", kEnc}, {"pic", kEnc}, {"thumb", kEnc},
		},
		hasIsActive: true, isActiveDefault: 0,
		junctions: []junction{versionJunctionThing},
		reqAdd:    []string{"name", "name_en"},
		reqMsg:    "名称和英文名称不能为空",
	},
	{
		seg: "version", table: "versions",
		cols: []col{
			{"code", kEnc}, {"name", kEnc}, {"name_en", kEnc}, {"description", kEnc}, {"sort_order", kNum},
		},
		hasIsActive: true, isActiveDefault: 1,
		reqAdd: []string{"code", "name"},
		reqMsg: "版本代码和名称不能为空",
	},
	{
		seg: "craft-tab", table: "craft_tabs",
		cols: []col{
			{"code", kEnc}, {"name", kEnc}, {"name_en", kEnc}, {"sort_order", kNum},
		},
		hasIsActive: true, isActiveDefault: 1,
		reqAdd: []string{"code", "name"},
		reqMsg: "标签代码和名称不能为空",
	},
}

// ---------------------------------------------------------------------------
// value handling
// ---------------------------------------------------------------------------

func colValue(m map[string]interface{}, cl col) interface{} {
	switch cl.kind {
	case kEnc:
		return apihelper.NullableEncoded(m, cl.name)
	case kNum:
		return apihelper.NullableNum(m, cl.name)
	default: // kRaw
		return apihelper.S(m, cl.name)
	}
}

// missingRequired reports whether any required field is empty. "version"/"tab"
// are validated as id lists (they arrive as a number or an array of numbers).
func missingRequired(m map[string]interface{}, fields []string) bool {
	for _, f := range fields {
		if f == "version" || f == "tab" {
			if len(apihelper.IntSlice(m, f)) == 0 {
				return true
			}
			continue
		}
		if strings.TrimSpace(apihelper.S(m, f)) == "" {
			return true
		}
	}
	return false
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

func idsToArgs(ids []int64) []interface{} {
	out := make([]interface{}, len(ids))
	for i, v := range ids {
		out[i] = v
	}
	return out
}

// ---------------------------------------------------------------------------
// version / tab relation helpers
// ---------------------------------------------------------------------------

func versionJunctionFor(table string) (junction, bool) {
	switch table {
	case "characters":
		return versionJunctionCharacter, true
	case "mobs":
		return versionJunctionMob, true
	case "plants":
		return versionJunctionPlant, true
	case "things":
		return versionJunctionThing, true
	case "crafts":
		return versionJunctionCraft, true
	}
	return junction{}, false
}

// batchEntityVersions returns id -> [{id,code,name,name_en}] for every entity id.
func batchEntityVersions(d *sql.DB, table string, ids []int64) map[int64][]gin.H {
	result := map[int64][]gin.H{}
	j, ok := versionJunctionFor(table)
	if !ok || len(ids) == 0 {
		return result
	}
	query := fmt.Sprintf(`SELECT j.%s AS eid, v.id, v.code, v.name, v.name_en
		FROM %s j INNER JOIN versions v ON j.version_id = v.id
		WHERE j.%s IN (%s)
		ORDER BY j.%s, j.is_primary DESC, v.sort_order ASC`,
		j.entityCol, j.table, j.entityCol, placeholders(len(ids)), j.entityCol)
	rows, err := db.QueryMaps(d, query, idsToArgs(ids)...)
	if err != nil {
		return result
	}
	for _, r := range rows {
		eid := asInt64(r["eid"])
		result[eid] = append(result[eid], gin.H{
			"id": r["id"], "code": r["code"], "name": r["name"], "name_en": r["name_en"],
		})
	}
	return result
}

// batchCraftTabs returns craft_id -> [{id,code,name,name_en}].
func batchCraftTabs(d *sql.DB, ids []int64) map[int64][]gin.H {
	result := map[int64][]gin.H{}
	if len(ids) == 0 {
		return result
	}
	query := fmt.Sprintf(`SELECT ctr.craft_id AS eid, t.id, t.code, t.name, t.name_en
		FROM craft_tab_relations ctr INNER JOIN craft_tabs t ON ctr.tab_id = t.id
		WHERE ctr.craft_id IN (%s)
		ORDER BY ctr.craft_id, ctr.is_primary DESC, t.sort_order ASC`, placeholders(len(ids)))
	rows, err := db.QueryMaps(d, query, idsToArgs(ids)...)
	if err != nil {
		return result
	}
	for _, r := range rows {
		eid := asInt64(r["eid"])
		result[eid] = append(result[eid], gin.H{
			"id": r["id"], "code": r["code"], "name": r["name"], "name_en": r["name_en"],
		})
	}
	return result
}

// enrichList adds version (and, for crafts, tab) arrays onto each row.
func enrichList(d *sql.DB, table string, list []map[string]interface{}, hasVersion, hasTab bool) {
	if len(list) == 0 {
		return
	}
	var ids []int64
	for _, row := range list {
		ids = append(ids, asInt64(row["id"]))
	}
	var versions map[int64][]gin.H
	var tabs map[int64][]gin.H
	if hasVersion {
		versions = batchEntityVersions(d, table, ids)
	}
	if hasTab && table == "crafts" {
		tabs = batchCraftTabs(d, ids)
	}
	for _, row := range list {
		id := asInt64(row["id"])
		if hasVersion {
			if v := versions[id]; v != nil {
				row["version"] = v
			} else {
				row["version"] = []gin.H{}
			}
		}
		if hasTab && table == "crafts" {
			if t := tabs[id]; t != nil {
				row["tab"] = t
			} else {
				row["tab"] = []gin.H{}
			}
		}
		delete(row, "version_name")
		delete(row, "version_name_en")
	}
}

func asInt64(v interface{}) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	}
	return 0
}

// ---------------------------------------------------------------------------
// list / info (public)
// ---------------------------------------------------------------------------

func listHandler(lc listConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var where []string
		var args []interface{}
		joins := ""

		if lc.hasVersion {
			if v := c.Query("version"); v != "" {
				if j, ok := versionJunctionFor(lc.table); ok {
					joins += fmt.Sprintf(" INNER JOIN %s ON %s.id = %s.%s INNER JOIN versions ON %s.version_id = versions.id",
						j.table, lc.table, j.table, j.entityCol, j.table)
					where = append(where, "versions.code = ?")
					args = append(args, v)
				}
			}
		}
		if lc.hasTab && lc.table == "crafts" {
			if tab := c.Query("tab"); tab != "" {
				joins += " INNER JOIN craft_tab_relations ctr ON crafts.id = ctr.craft_id INNER JOIN craft_tabs ct ON ctr.tab_id = ct.id"
				where = append(where, "ct.code = ?")
				args = append(args, tab)
			}
		}
		if kw := c.Query("keyword"); kw != "" && len(lc.search) > 0 {
			enc := util.UnicodeEncode(kw)
			var parts []string
			for _, f := range lc.search {
				parts = append(parts, fmt.Sprintf("%s.%s LIKE ? ESCAPE '/'", lc.table, f))
				args = append(args, "%"+enc+"%")
			}
			where = append(where, "("+strings.Join(parts, " OR ")+")")
		}
		if c.Query("all") != "1" {
			where = append(where, fmt.Sprintf("%s.is_active = 1", lc.table))
		}
		whereClause := ""
		if len(where) > 0 {
			whereClause = "WHERE " + strings.Join(where, " AND ")
		}
		orderClause := ""
		if lc.orderBy != "" {
			orderClause = " ORDER BY " + lc.orderBy
		}

		d, err := db.Open(dbName)
		if err != nil {
			response.Error(c, err.Error(), err.Error())
			return
		}

		pageNo := c.Query("pageNo")
		pageSize := c.Query("pageSize")
		if pageNo != "" && pageSize != "" {
			pn := util.AtoiDefault(pageNo, 1)
			ps := util.AtoiDefault(pageSize, 10)
			start := 0
			if pn > 1 {
				start = (pn - 1) * ps
			}
			listQuery := fmt.Sprintf("SELECT DISTINCT %s.* FROM %s%s %s%s LIMIT ?, ?", lc.table, lc.table, joins, whereClause, orderClause)
			listArgs := append(append([]interface{}{}, args...), start, ps)
			list, err := db.QueryMaps(d, listQuery, listArgs...)
			if err != nil {
				response.Error(c, err.Error(), err.Error())
				return
			}
			enrichList(d, lc.table, list, lc.hasVersion, lc.hasTab)
			countRow, err := db.QueryMap(d, fmt.Sprintf("SELECT COUNT(DISTINCT %s.id) AS sum FROM %s%s %s", lc.table, lc.table, joins, whereClause), args...)
			if err != nil {
				response.Error(c, err.Error(), err.Error())
				return
			}
			var total interface{} = 0
			if countRow != nil {
				total = countRow["sum"]
			}
			response.Success(c, gin.H{
				"list": list,
				"pager": gin.H{
					"pageSize": ps,
					"pageNo":   pn,
					"total":    total,
				},
			}, "请求成功")
			return
		}

		list, err := db.QueryMaps(d, fmt.Sprintf("SELECT DISTINCT %s.* FROM %s%s %s%s", lc.table, lc.table, joins, whereClause, orderClause), args...)
		if err != nil {
			response.Error(c, err.Error(), err.Error())
			return
		}
		enrichList(d, lc.table, list, lc.hasVersion, lc.hasTab)
		response.Success(c, list, "请求成功")
	}
}

func infoHandler(ic infoConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")
		d, err := db.Open(dbName)
		if err != nil {
			response.Error(c, err.Error(), err.Error())
			return
		}
		row, err := db.QueryMap(d, fmt.Sprintf("SELECT %s.* FROM %s WHERE %s.id = ?", ic.table, ic.table, ic.table), id)
		if err != nil {
			response.Error(c, err.Error(), err.Error())
			return
		}
		if row == nil {
			response.Error(c, "", "无数据")
			return
		}
		eid := asInt64(row["id"])
		if ic.hasVersion {
			row["version"] = batchEntityVersions(d, ic.table, []int64{eid})[eid]
			if row["version"] == nil {
				row["version"] = []gin.H{}
			}
		}
		if ic.hasTab && ic.table == "crafts" {
			row["tab"] = batchCraftTabs(d, []int64{eid})[eid]
			if row["tab"] == nil {
				row["tab"] = []gin.H{}
			}
		}
		delete(row, "version_name")
		delete(row, "version_name_en")
		response.Success(c, row, "请求成功")
	}
}

// ---------------------------------------------------------------------------
// mutations (auth required)
// ---------------------------------------------------------------------------

func addHandler(e entity) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, msg := middleware.VerifyAuthorization(c)
		if msg != "" {
			response.Error(c, "", msg)
			return
		}
		m := apihelper.Body(c)
		if missingRequired(m, e.reqAdd) {
			response.Error(c, "", e.reqMsg)
			return
		}

		names := make([]string, 0, len(e.cols)+1)
		phs := make([]string, 0, len(e.cols)+1)
		args := make([]interface{}, 0, len(e.cols)+1)
		for _, cl := range e.cols {
			names = append(names, cl.name)
			phs = append(phs, "?")
			args = append(args, colValue(m, cl))
		}
		if e.hasIsActive {
			names = append(names, "is_active")
			phs = append(phs, "?")
			args = append(args, apihelper.IsActive(m, e.isActiveDefault))
		}
		query := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", e.table, strings.Join(names, ", "), strings.Join(phs, ", "))

		d, err := db.Open(dbName)
		if err != nil {
			response.Error(c, err.Error(), err.Error())
			return
		}

		// Fast path: no relations to persist.
		if len(e.junctions) == 0 {
			apihelper.OperateReturnID(c, d, user.UID, e.table, "添加", query, args...)
			return
		}

		tx, err := d.Begin()
		if err != nil {
			response.Error(c, err.Error(), "添加失败")
			return
		}
		res, err := tx.Exec(query, args...)
		if err != nil {
			_ = tx.Rollback()
			response.Error(c, err.Error(), err.Error())
			return
		}
		newID, _ := res.LastInsertId()
		for _, j := range e.junctions {
			ids := apihelper.IntSlice(m, j.field)
			for idx, refID := range ids {
				primary := 0
				if idx == 0 {
					primary = 1
				}
				_, err := tx.Exec(
					fmt.Sprintf("INSERT INTO %s(%s, %s, is_primary) VALUES(?, ?, ?)", j.table, j.entityCol, j.refCol),
					newID, refID, primary)
				if err != nil {
					_ = tx.Rollback()
					response.Error(c, err.Error(), "添加关系失败")
					return
				}
			}
		}
		if err := tx.Commit(); err != nil {
			response.Error(c, err.Error(), "添加失败")
			return
		}
		util.UpdateUserLastLoginTime(user.UID)
		response.Success(c, gin.H{"id": newID}, "添加成功")
	}
}

func modifyHandler(e entity) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, msg := middleware.VerifyAuthorization(c)
		if msg != "" {
			response.Error(c, "", msg)
			return
		}
		m := apihelper.Body(c)
		id := apihelper.I(m, "id")
		if id == 0 {
			response.Error(c, "", "ID不能为空")
			return
		}
		if len(e.reqModify) > 0 && missingRequired(m, e.reqModify) {
			response.Error(c, "", e.reqMsg)
			return
		}

		sets := make([]string, 0, len(e.cols)+1)
		args := make([]interface{}, 0, len(e.cols)+2)
		for _, cl := range e.cols {
			sets = append(sets, cl.name+" = ?")
			args = append(args, colValue(m, cl))
		}
		if e.hasIsActive {
			sets = append(sets, "is_active = ?")
			args = append(args, apihelper.IsActive(m, e.isActiveDefault))
		}
		args = append(args, id)
		query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", e.table, strings.Join(sets, ", "))

		d, err := db.Open(dbName)
		if err != nil {
			response.Error(c, err.Error(), err.Error())
			return
		}

		// Determine which relations were supplied in the request body.
		var active []junction
		for _, j := range e.junctions {
			if apihelper.Has(m, j.field) {
				active = append(active, j)
			}
		}
		if len(active) == 0 {
			apihelper.OperateNoReturn(c, d, user.UID, e.table, "修改", query, args...)
			return
		}

		tx, err := d.Begin()
		if err != nil {
			response.Error(c, err.Error(), "修改失败")
			return
		}
		if _, err := tx.Exec(query, args...); err != nil {
			_ = tx.Rollback()
			response.Error(c, err.Error(), err.Error())
			return
		}
		for _, j := range active {
			if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s = ?", j.table, j.entityCol), id); err != nil {
				_ = tx.Rollback()
				response.Error(c, err.Error(), "修改关系失败")
				return
			}
			ids := apihelper.IntSlice(m, j.field)
			for idx, refID := range ids {
				primary := 0
				if idx == 0 {
					primary = 1
				}
				if _, err := tx.Exec(
					fmt.Sprintf("INSERT INTO %s(%s, %s, is_primary) VALUES(?, ?, ?)", j.table, j.entityCol, j.refCol),
					id, refID, primary); err != nil {
					_ = tx.Rollback()
					response.Error(c, err.Error(), "修改关系失败")
					return
				}
			}
		}
		if err := tx.Commit(); err != nil {
			response.Error(c, err.Error(), "修改失败")
			return
		}
		util.UpdateUserLastLoginTime(user.UID)
		response.Success(c, nil, "修改成功")
	}
}

func deleteHandler(e entity) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, msg := middleware.VerifyAuthorization(c)
		if msg != "" {
			response.Error(c, "", msg)
			return
		}
		m := apihelper.Body(c)
		var idVal interface{}
		if bodyID := apihelper.I(m, "id"); bodyID != 0 {
			idVal = bodyID
		} else if q := c.Query("id"); q != "" {
			idVal = q
		} else {
			response.Error(c, "", "ID不能为空")
			return
		}
		d, err := db.Open(dbName)
		if err != nil {
			response.Error(c, err.Error(), err.Error())
			return
		}
		// ON DELETE CASCADE removes the linked version / tab relations.
		apihelper.OperateNoReturn(c, d, user.UID, e.table, "删除", fmt.Sprintf("DELETE FROM %s WHERE id = ?", e.table), idVal)
	}
}

package applecatalog

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

//go:embed seed.json
var seedFS embed.FS

const dbName = db.Diary

var (
	seedMu     sync.Mutex
	seeded     bool
	repairMu   sync.Mutex
	repaired   bool
	fillMu     sync.Mutex
	filled     bool
	deviceMu   sync.Mutex
	deviceFill bool
)

func Register(r *gin.RouterGroup) {
	r.Use(openCORS)
	r.GET("", handleAll)
	r.GET("/", handleAll)
	r.POST("/device", handleSaveDevice)
	r.DELETE("/device", handleDeleteDevice)
	r.POST("/chip", handleSaveChip)
	r.DELETE("/chip", handleDeleteChip)
}

func openCORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
	c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(204)
		return
	}
	c.Next()
}

func handleAll(c *gin.Context) {
	_ = SeedIfEmpty()
	_ = RepairChipFlatFields()
	_ = FillMissingChipSpecs()
	_ = FillMissingDeviceSpecs()
	full := c.Query("full") == "1"
	diary, err := db.Open(dbName)
	if err != nil {
		response.Error(c, "", "数据库未连接")
		return
	}
	devices, err := loadRows(diary, "apple_devices", full)
	if err != nil {
		response.Error(c, err.Error(), "读取设备失败")
		return
	}
	chips, err := loadRows(diary, "apple_chips", full)
	if err != nil {
		response.Error(c, err.Error(), "读取芯片失败")
		return
	}
	response.Success(c, gin.H{"devices": devices, "chips": chips}, "")
}

func handleSaveDevice(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	body, ok := bindBody(c)
	if !ok {
		return
	}
	if err := saveRow("apple_devices", body, true); err != nil {
		response.Error(c, err.Error(), "保存设备失败")
		return
	}
	response.Success(c, gin.H{"id": asString(body["id"])}, "已保存")
}

func handleDeleteDevice(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	deleteRow(c, "apple_devices")
}

func handleSaveChip(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	body, ok := bindBody(c)
	if !ok {
		return
	}
	if err := saveRow("apple_chips", body, false); err != nil {
		response.Error(c, err.Error(), "保存芯片失败")
		return
	}
	response.Success(c, gin.H{"id": asString(body["id"])}, "已保存")
}

func handleDeleteChip(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	deleteRow(c, "apple_chips")
}

func bindBody(c *gin.Context) (map[string]interface{}, bool) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(asString(body["name"])) == "" {
		response.Error(c, "", "缺少名称")
		return nil, false
	}
	return body, true
}

func deleteRow(c *gin.Context, table string) {
	var body struct {
		ID string `json:"id"`
	}
	_ = c.ShouldBindJSON(&body)
	if strings.TrimSpace(body.ID) == "" {
		response.Error(c, "", "缺少 id")
		return
	}
	diary, err := db.Open(dbName)
	if err != nil {
		response.Error(c, "", "数据库未连接")
		return
	}
	if _, err = diary.Exec(`DELETE FROM `+table+` WHERE id = ?`, body.ID); err != nil {
		response.Error(c, err.Error(), "删除失败")
		return
	}
	response.Success(c, nil, "已删除")
}

func requireAdmin(c *gin.Context) bool {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return false
	}
	if user == nil || !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return false
	}
	return true
}

func loadRows(diary *sql.DB, table string, full bool) ([]map[string]interface{}, error) {
	rows, err := db.QueryMaps(diary, `SELECT record_json, source_json FROM `+table+` ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := decodeRecord(asString(row["record_json"]), asString(row["source_json"]), full)
		if item != nil {
			out = append(out, item)
		}
	}
	return out, nil
}

func decodeRecord(record string, source string, full bool) map[string]interface{} {
	if record == "" {
		return nil
	}
	var item map[string]interface{}
	if err := json.Unmarshal([]byte(record), &item); err != nil {
		return nil
	}
	// 旧数据扁平字段可能缺 iceRate，读的时候先从 source 补上。
	fillChipFromSource(item, source)
	if full && source != "" {
		var raw interface{}
		if err := json.Unmarshal([]byte(source), &raw); err == nil {
			item["source"] = raw
		}
	}
	return item
}

// 把日记时代的嵌套 CPU/GPU 抄回展示字段，缺啥补啥，有值不覆盖。
func fillChipFromSource(item map[string]interface{}, sourceJSON string) bool {
	if sourceJSON == "" || item == nil {
		return false
	}
	var source map[string]interface{}
	if err := json.Unmarshal([]byte(sourceJSON), &source); err != nil {
		return false
	}
	cpu := firstMap(source["cpu"])
	gpu := firstMap(source["gpu"])
	fire := asMap(cpu["fire"])
	ice := asMap(cpu["ice"])
	ai := asMap(source["ai"])
	changed := false
	changed = setIfEmpty(item, "fireCore", fire["core"]) || changed
	changed = setIfEmpty(item, "iceCore", ice["core"]) || changed
	changed = setIfEmpty(item, "fireRate", fire["rate"]) || changed
	changed = setIfEmpty(item, "iceRate", ice["rate"]) || changed
	changed = setIfEmpty(item, "gpuCore", gpu["core"]) || changed
	changed = setIfEmpty(item, "gpuInfo", gpu["info"]) || changed
	changed = setIfEmpty(item, "gpuInfo", gpu["brand"]) || changed
	changed = setIfEmpty(item, "aiCore", ai["core"]) || changed
	changed = setIfEmpty(item, "aiRate", ai["rate"]) || changed
	changed = setIfEmpty(item, "tech", source["tech"]) || changed
	changed = setIfEmpty(item, "die", source["dieSize"]) || changed
	changed = setIfEmpty(item, "isa", source["isa"]) || changed
	changed = setIfEmpty(item, "model", source["model"]) || changed
	changed = setIfEmpty(item, "transistor", source["transistorCount"]) || changed
	return changed
}

func firstMap(v interface{}) map[string]interface{} {
	switch val := v.(type) {
	case []interface{}:
		if len(val) == 0 {
			return map[string]interface{}{}
		}
		return asMap(val[0])
	default:
		return asMap(v)
	}
}

func asMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func setIfEmpty(item map[string]interface{}, key string, value interface{}) bool {
	if value == nil {
		return false
	}
	if !isEmptyValue(item[key]) {
		return false
	}
	if isEmptyValue(value) {
		return false
	}
	item[key] = value
	return true
}

func isEmptyValue(v interface{}) bool {
	switch val := v.(type) {
	case nil:
		return true
	case string:
		s := strings.TrimSpace(val)
		return s == "" || s == "-" || s == "0"
	case float64:
		return val == 0
	case int:
		return val == 0
	case int64:
		return val == 0
	case json.Number:
		n, err := val.Float64()
		return err != nil || n == 0
	default:
		if arr, ok := v.([]interface{}); ok {
			return len(arr) == 0
		}
		return false
	}
}

// RepairChipFlatFields 把 source_json 里的主频等写回 record_json，只跑一次。
func RepairChipFlatFields() error {
	repairMu.Lock()
	defer repairMu.Unlock()
	if repaired {
		return nil
	}
	err := repairChipFlatFields()
	if err == nil {
		repaired = true
	}
	return err
}

func repairChipFlatFields() error {
	diary, err := db.Open(dbName)
	if err != nil {
		return err
	}
	rows, err := db.QueryMaps(diary, `SELECT id, record_json, source_json FROM apple_chips`)
	if err != nil {
		return err
	}
	now := util.NowString()
	for _, row := range rows {
		recordText := asString(row["record_json"])
		sourceText := asString(row["source_json"])
		if recordText == "" || sourceText == "" {
			continue
		}
		var item map[string]interface{}
		if err = json.Unmarshal([]byte(recordText), &item); err != nil {
			continue
		}
		if !fillChipFromSource(item, sourceText) {
			continue
		}
		raw, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if _, err = diary.Exec(`UPDATE apple_chips SET record_json=?, date_modify=? WHERE id=?`,
			string(raw), now, asString(row["id"])); err != nil {
			return err
		}
	}
	return nil
}

// FillMissingChipSpecs 用种子补空字段，并插入库里还没有的芯片。
func FillMissingChipSpecs() error {
	fillMu.Lock()
	defer fillMu.Unlock()
	if filled {
		return nil
	}
	err := fillMissingChipSpecs()
	if err == nil {
		filled = true
	}
	return err
}

func fillMissingChipSpecs() error {
	raw, err := seedFS.ReadFile("seed.json")
	if err != nil {
		return err
	}
	var file seedFile
	if err = json.Unmarshal(raw, &file); err != nil {
		return err
	}
	byID := map[string]map[string]interface{}{}
	byName := map[string]map[string]interface{}{}
	for _, chip := range file.Chips {
		id := asString(chip["id"])
		if id != "" {
			byID[id] = chip
		}
		key := asString(chip["series"]) + ":" + asString(chip["name"])
		if byName[key] == nil {
			byName[key] = chip
		}
	}
	diary, err := db.Open(dbName)
	if err != nil {
		return err
	}
	rows, err := db.QueryMaps(diary, `SELECT id, record_json, source_json FROM apple_chips`)
	if err != nil {
		return err
	}
	present := map[string]bool{}
	presentVar := map[string]bool{}
	now := util.NowString()
	for _, row := range rows {
		id := asString(row["id"])
		present[id] = true
		recordText := asString(row["record_json"])
		if recordText == "" {
			continue
		}
		var item map[string]interface{}
		if err = json.Unmarshal([]byte(recordText), &item); err != nil {
			continue
		}
		presentVar[chipVariantKey(item)] = true
		seed := byID[id]
		if seed == nil {
			seed = byName[asString(item["series"])+":"+asString(item["name"])]
		}
		changed := fillChipFromSource(item, asString(row["source_json"]))
		if seed != nil {
			for _, key := range []string{
				"model", "tech", "die", "isa", "fireCore", "iceCore", "fireRate", "iceRate",
				"gpuCore", "gpuInfo", "aiCore", "aiRate", "transistor", "release", "devices",
			} {
				changed = setIfEmpty(item, key, seed[key]) || changed
			}
			// A16 表上是 4nm；A19 面积曾误写成和 Pro 一样
			if asString(item["name"]) == "A16 Bionic" && asInt(item["tech"]) == 5 && seed["tech"] != nil {
				item["tech"] = seed["tech"]
				changed = true
			}
			if asString(item["name"]) == "A19" && seed["die"] != nil {
				item["die"] = seed["die"]
				changed = true
			}
		}
		if !changed {
			continue
		}
		next, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if _, err = diary.Exec(`UPDATE apple_chips SET record_json=?, date_modify=? WHERE id=?`,
			string(next), now, id); err != nil {
			return err
		}
	}
	var sort int
	_ = diary.QueryRow(`SELECT COALESCE(MAX(sort_order), 0) FROM apple_chips`).Scan(&sort)
	tx, err := diary.Begin()
	if err != nil {
		return err
	}
	inserted := 0
	for _, chip := range file.Chips {
		id := asString(chip["id"])
		if id == "" || present[id] || presentVar[chipVariantKey(chip)] {
			continue
		}
		sort++
		clone, err := cloneMap(chip)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if err = insertSeed(tx, "apple_chips", clone, sort, false); err != nil {
			_ = tx.Rollback()
			return err
		}
		inserted++
	}
	if inserted == 0 {
		_ = tx.Rollback()
		return nil
	}
	return tx.Commit()
}

// FillMissingDeviceSpecs 插入种子里还没有的设备，空字段用种子补。
func FillMissingDeviceSpecs() error {
	deviceMu.Lock()
	defer deviceMu.Unlock()
	if deviceFill {
		return nil
	}
	err := fillMissingDeviceSpecs()
	if err == nil {
		deviceFill = true
	}
	return err
}

func fillMissingDeviceSpecs() error {
	raw, err := seedFS.ReadFile("seed.json")
	if err != nil {
		return err
	}
	var file seedFile
	if err = json.Unmarshal(raw, &file); err != nil {
		return err
	}
	byID := map[string]map[string]interface{}{}
	for _, device := range file.Devices {
		id := asString(device["id"])
		if id != "" {
			byID[id] = device
		}
	}
	diary, err := db.Open(dbName)
	if err != nil {
		return err
	}
	rows, err := db.QueryMaps(diary, `SELECT id, record_json, source_json FROM apple_devices`)
	if err != nil {
		return err
	}
	present := map[string]bool{}
	presentName := map[string]bool{}
	now := util.NowString()
	for _, row := range rows {
		id := asString(row["id"])
		present[id] = true
		recordText := asString(row["record_json"])
		if recordText == "" {
			continue
		}
		var item map[string]interface{}
		if err = json.Unmarshal([]byte(recordText), &item); err != nil {
			continue
		}
		presentName[asString(item["family"])+":"+asString(item["name"])] = true
		seed := byID[id]
		if seed == nil {
			continue
		}
		changed := false
		keys := []string{
			"chip", "cpuRate", "cpuCore", "cpuBit", "memoryGb", "memoryType", "battery", "batteryUnit",
			"material", "water", "duration", "caseMm", "dust", "cellular", "wifi", "bluetooth", "uwb",
			"heart", "gps", "accel", "gyro", "altimeter", "osFrom", "osTo", "unsupported", "require",
			"chargeTime", "port", "image", "connect", "charge", "unlock", "pencil", "hardware", "models",
		}
		for _, key := range keys {
			changed = setIfEmpty(item, key, seed[key]) || changed
		}
		if asString(item["family"]) == "watch" && seed["osTo"] != nil && asString(item["osTo"]) == "9.3.1" {
			item["osTo"] = seed["osTo"]
			changed = true
		}
		if !changed {
			continue
		}
		next, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if _, err = diary.Exec(`UPDATE apple_devices SET record_json=?, date_modify=? WHERE id=?`,
			string(next), now, id); err != nil {
			return err
		}
	}
	var sort int
	_ = diary.QueryRow(`SELECT COALESCE(MAX(sort_order), 0) FROM apple_devices`).Scan(&sort)
	tx, err := diary.Begin()
	if err != nil {
		return err
	}
	inserted := 0
	for _, device := range file.Devices {
		id := asString(device["id"])
		if id == "" || present[id] || presentName[asString(device["family"])+":"+asString(device["name"])] {
			continue
		}
		sort++
		clone, err := cloneMap(device)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		if err = insertSeed(tx, "apple_devices", clone, sort, true); err != nil {
			_ = tx.Rollback()
			return err
		}
		inserted++
	}
	if inserted == 0 {
		_ = tx.Rollback()
		return nil
	}
	return tx.Commit()
}

func chipVariantKey(item map[string]interface{}) string {
	return asString(item["series"]) + ":" + asString(item["name"]) + ":" + asString(item["model"]) + ":" + fmt.Sprint(asInt(item["die"])) + ":" + asString(item["tech"])
}

func cloneMap(item map[string]interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	err = json.Unmarshal(raw, &out)
	return out, err
}

func saveRow(table string, body map[string]interface{}, device bool) error {
	source, hasSource := body["source"]
	delete(body, "source")
	if asString(body["id"]) == "" {
		family := asString(body["family"])
		if family == "" {
			family = strings.ToLower(asString(body["series"]))
		}
		body["id"] = slug(family, asString(body["name"]))
	}
	if asInt(body["year"]) == 0 {
		body["year"] = yearOf(asString(body["release"]))
	}
	record, err := json.Marshal(body)
	if err != nil {
		return err
	}
	var sourceText interface{}
	if hasSource && source != nil {
		raw, err := json.Marshal(source)
		if err != nil {
			return err
		}
		sourceText = string(raw)
	}
	diary, err := db.Open(dbName)
	if err != nil {
		return err
	}
	var count int
	if err = diary.QueryRow(`SELECT COUNT(1) FROM `+table+` WHERE id = ?`, asString(body["id"])).Scan(&count); err != nil {
		return err
	}
	now := util.NowString()
	if count == 0 {
		var sort int
		_ = diary.QueryRow(`SELECT COALESCE(MAX(sort_order), 0) + 1 FROM ` + table).Scan(&sort)
		_, err = diary.Exec(`INSERT INTO `+table+`
			(id, family, series, name, name_short, release_date, year, sort_order, record_json, source_json, date_modify)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
			asString(body["id"]), familyOf(body, device), asString(body["series"]), asString(body["name"]),
			shortName(body), asString(body["release"]), asInt(body["year"]), sort, string(record), sourceText, now)
		return err
	}
	_, err = diary.Exec(`UPDATE `+table+` SET family=?, series=?, name=?, name_short=?, release_date=?, year=?, record_json=?, date_modify=? WHERE id=?`,
		familyOf(body, device), asString(body["series"]), asString(body["name"]), shortName(body),
		asString(body["release"]), asInt(body["year"]), string(record), now, asString(body["id"]))
	if err != nil || !hasSource {
		return err
	}
	_, err = diary.Exec(`UPDATE `+table+` SET source_json=? WHERE id=?`, sourceText, asString(body["id"]))
	return err
}

// 芯片没有产品族，family 列记系列字母，和设备表共用同一套写入。
func familyOf(body map[string]interface{}, device bool) string {
	if device {
		if family := asString(body["family"]); family != "" {
			return family
		}
		return "iphone"
	}
	return strings.ToUpper(asString(body["series"]))
}

func shortName(body map[string]interface{}) string {
	if name := asString(body["nameShort"]); name != "" {
		return name
	}
	return asString(body["name"])
}

type seedFile struct {
	Devices []map[string]interface{} `json:"devices"`
	Chips   []map[string]interface{} `json:"chips"`
}

// SeedIfEmpty 只在两张表都是空的时候写入种子，已有编辑不再覆盖。
func SeedIfEmpty() error {
	seedMu.Lock()
	defer seedMu.Unlock()
	if seeded {
		return nil
	}
	err := seedIfEmpty()
	if err == nil {
		seeded = true
	}
	return err
}

func seedIfEmpty() error {
	diary, err := db.Open(dbName)
	if err != nil {
		return err
	}
	var devices, chips int
	if err = diary.QueryRow(`SELECT COUNT(1) FROM apple_devices`).Scan(&devices); err != nil {
		return err
	}
	if err = diary.QueryRow(`SELECT COUNT(1) FROM apple_chips`).Scan(&chips); err != nil {
		return err
	}
	if devices > 0 || chips > 0 {
		return nil
	}
	raw, err := seedFS.ReadFile("seed.json")
	if err != nil {
		return err
	}
	var file seedFile
	if err = json.Unmarshal(raw, &file); err != nil {
		return err
	}
	tx, err := diary.Begin()
	if err != nil {
		return err
	}
	for i, item := range file.Devices {
		if err = insertSeed(tx, "apple_devices", item, i, true); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	for i, item := range file.Chips {
		if err = insertSeed(tx, "apple_chips", item, i, false); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func insertSeed(tx *sql.Tx, table string, item map[string]interface{}, order int, device bool) error {
	source := item["source"]
	delete(item, "source")
	record, err := json.Marshal(item)
	if err != nil {
		return err
	}
	var sourceText interface{}
	if source != nil {
		raw, err := json.Marshal(source)
		if err != nil {
			return err
		}
		sourceText = string(raw)
	}
	_, err = tx.Exec(`INSERT INTO `+table+`
		(id, family, series, name, name_short, release_date, year, sort_order, record_json, source_json, date_modify)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		asString(item["id"]), familyOf(item, device), asString(item["series"]), asString(item["name"]),
		shortName(item), asString(item["release"]), asInt(item["year"]), order, string(record), sourceText, util.NowString())
	return err
}

var nonSlug = regexp.MustCompile(`[^\w.\p{Han}-]+`)

func slug(family, name string) string {
	text := strings.ToLower(strings.TrimSpace(name))
	text = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '-'
		}
		return r
	}, text)
	text = nonSlug.ReplaceAllString(text, "")
	if family == "" {
		return text
	}
	return family + "-" + text
}

func yearOf(date string) int {
	if len(date) < 4 {
		return 0
	}
	var year int
	_, _ = fmt.Sscanf(date, "%d", &year)
	if year < 1990 || year > 2100 {
		return 0
	}
	return year
}

func asString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(val))
	}
}

func asInt(v interface{}) int {
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	case json.Number:
		n, _ := val.Int64()
		return int(n)
	case string:
		var n int
		_, _ = fmt.Sscanf(val, "%d", &n)
		return n
	default:
		return 0
	}
}

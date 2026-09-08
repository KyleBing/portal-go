package apihelper

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

// normalizeVal converts driver values into JSON-friendly Go values.
func normalizeVal(v interface{}) interface{} {
	switch t := v.(type) {
	case []byte:
		return string(t)
	case time.Time:
		return util.DateFormatter(t, "")
	default:
		return v
	}
}

// QueryMaps runs a query and returns rows as generic maps (column -> value).
func QueryMaps(database *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	out := []map[string]interface{}{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		m := map[string]interface{}{}
		for i, c := range cols {
			m[c] = normalizeVal(vals[i])
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// QueryMap runs a query and returns the first row as a map, or nil if none.
func QueryMap(database *sql.DB, query string, args ...interface{}) (map[string]interface{}, error) {
	rows, err := QueryMaps(database, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// MapInt reads an integer field from a result map.
func MapInt(m map[string]interface{}, key string) int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case []byte:
		n, _ := strconv.ParseInt(string(t), 10, 64)
		return n
	case string:
		n, _ := strconv.ParseInt(t, 10, 64)
		return n
	}
	return 0
}

// MapStr reads a string field from a result map.
func MapStr(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// Body reads the JSON request body into a generic map, mirroring express's
// req.body access pattern. It caches the parsed body on the context so it can
// be read multiple times within a handler.
func Body(c *gin.Context) map[string]interface{} {
	if cached, ok := c.Get("parsedBody"); ok {
		if m, ok := cached.(map[string]interface{}); ok {
			return m
		}
	}
	m := map[string]interface{}{}
	_ = c.ShouldBindJSON(&m)
	c.Set("parsedBody", m)
	return m
}

// S returns the string value for key (empty string if missing/null).
func S(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(v)
	}
}

// Has reports whether the key exists in the body (even if null), like
// req.body.hasOwnProperty.
func Has(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

// Present reports whether the key exists and is not null.
func Present(m map[string]interface{}, key string) bool {
	v, ok := m[key]
	return ok && v != nil
}

// I returns the integer value for key (0 if missing/invalid).
func I(m map[string]interface{}, key string) int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return int64(t)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		return n
	case bool:
		if t {
			return 1
		}
		return 0
	}
	return 0
}

// F returns the float value for key.
func F(m map[string]interface{}, key string) float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return t
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f
	}
	return 0
}

// EncodedStr returns unicodeEncode(value) for direct storage as a parameter.
func EncodedStr(m map[string]interface{}, key string) string {
	return util.UnicodeEncode(S(m, key))
}

// NullableEncoded mirrors processStringValue: returns nil (SQL NULL) for
// empty/whitespace-only values, otherwise the unicode-encoded string.
func NullableEncoded(m map[string]interface{}, key string) interface{} {
	v := S(m, key)
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return util.UnicodeEncode(v)
}

// NullableNum mirrors processNumberValue: nil for empty, else the number.
func NullableNum(m map[string]interface{}, key string) interface{} {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	if s, isStr := v.(string); isStr {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		return s
	}
	return v
}

// IsActive mirrors: req.body.is_active !== undefined ? (is_active?1:0) : def
func IsActive(m map[string]interface{}, def int) int {
	raw, exists := m["is_active"]
	if !exists {
		return def
	}
	switch t := raw.(type) {
	case bool:
		if t {
			return 1
		}
		return 0
	case float64:
		if t != 0 {
			return 1
		}
		return 0
	case string:
		if t != "" && t != "0" && t != "false" {
			return 1
		}
		return 0
	case nil:
		return 0
	}
	return def
}

// IntSlice parses a value that may be a single number or an array of numbers
// (used for starve-new version/tab relations).
func IntSlice(m map[string]interface{}, k string) []int64 {
	v, ok := m[k]
	if !ok || v == nil {
		return nil
	}
	var out []int64
	switch t := v.(type) {
	case []interface{}:
		for _, item := range t {
			if item == nil {
				continue
			}
			switch n := item.(type) {
			case float64:
				out = append(out, int64(n))
			case string:
				if x, err := strconv.ParseInt(strings.TrimSpace(n), 10, 64); err == nil {
					out = append(out, x)
				}
			}
		}
	case float64:
		out = append(out, int64(t))
	case string:
		if x, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64); err == nil {
			out = append(out, x)
		}
	}
	return out
}

// OperateReturnID runs an INSERT and returns {id: insertId}, mirroring
// operate_db_and_return_added_id.
func OperateReturnID(c *gin.Context, database *sql.DB, uid int64, dataName, operation, query string, args ...interface{}) {
	res, err := database.Exec(query, args...)
	if err != nil {
		response.Error(c, err.Error(), dataName+operation+"失败")
		return
	}
	id, _ := res.LastInsertId()
	util.UpdateUserLastLoginTime(uid)
	response.Success(c, gin.H{"id": id}, operation+"成功")
}

// OperateNoReturn runs a statement and returns success with no data, mirroring
// operate_db_without_return.
func OperateNoReturn(c *gin.Context, database *sql.DB, uid int64, dataName, operation, query string, args ...interface{}) {
	_, err := database.Exec(query, args...)
	if err != nil {
		response.Error(c, err.Error(), dataName+operation+"失败")
		return
	}
	util.UpdateUserLastLoginTime(uid)
	response.Success(c, nil, operation+"成功")
}

// PageStart computes the SQL OFFSET from pageNo/pageSize.
func PageStart(pageNo, pageSize int) int {
	if pageNo < 1 {
		pageNo = 1
	}
	return (pageNo - 1) * pageSize
}

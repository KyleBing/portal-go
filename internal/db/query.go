package db

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

// QueryMaps runs a query and returns every row as a generic map keyed by column
// name. This mirrors the dynamic `SELECT *` behaviour of the original Node/Express
// implementation which returned plain JSON objects.
func QueryMaps(d *sql.DB, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := d.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMaps(rows)
}

// QueryMap runs a query and returns the first row as a generic map, or nil when
// there are no rows (matching the Node helper's `isSingleValue` behaviour).
func QueryMap(d *sql.DB, query string, args ...interface{}) (map[string]interface{}, error) {
	list, err := QueryMaps(d, query, args...)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

func scanMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	colTypes, _ := rows.ColumnTypes()
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
		m := make(map[string]interface{}, len(cols))
		for i, c := range cols {
			var typeName string
			if colTypes != nil && i < len(colTypes) {
				typeName = colTypes[i].DatabaseTypeName()
			}
			m[c] = convertValue(vals[i], typeName)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func convertValue(v interface{}, dbType string) interface{} {
	switch val := v.(type) {
	case nil:
		return nil
	case []byte:
		// BIT columns (e.g. is_active in starve_advance) come back as raw bytes.
		// Convert them to an integer so the JSON output matches the original
		// Node implementation which coerced the Buffer to a number.
		if strings.ToUpper(dbType) == "BIT" {
			if len(val) > 0 {
				return int64(val[0])
			}
			return int64(0)
		}
		s := string(val)
		return coerceString(s, dbType)
	case string:
		return coerceString(val, dbType)
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	default:
		return val
	}
}

func coerceString(s, dbType string) interface{} {
	t := strings.ToUpper(dbType)
	switch t {
	case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT", "YEAR":
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
	case "FLOAT", "DOUBLE", "DECIMAL", "NEWDECIMAL":
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	}
	return s
}

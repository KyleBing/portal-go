package billimport

import (
	"database/sql"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/KyleBing/portal-go/internal/setup"
	"github.com/KyleBing/portal-go/internal/util"
)

func readFileBytes(relPath string) ([]byte, error) {
	abs := filepath.Join(setup.ProjectRoot(), relPath)
	return os.ReadFile(abs)
}

type contentRow struct {
	Name  string
	Price *float64
	Raw   string
}

func parseContent(content string) []contentRow {
	var parsed []contentRow
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		row := contentRow{Name: parts[0], Raw: line}
		if len(parts) >= 2 {
			if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
				row.Price = &v
			}
		}
		parsed = append(parsed, row)
	}
	return parsed
}

func findBillDiary(uid int64, date string) (map[string]interface{}, error) {
	diary, err := openDB()
	if err != nil {
		return nil, err
	}
	start := date + " 00:00:00"
	end := date + " 23:59:59"
	rows, err := diary.Query(`
		SELECT id, date, title, content, weather, temperature, temperature_outside, is_public, is_markdown
		FROM diaries
		WHERE uid=? AND category='bill' AND date >= ? AND date <= ?
		ORDER BY id ASC`, uid, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id                                   int64
			dateVal                              time.Time
			title, content, weather              string
			temp, tempOut                        float64
			isPublic, isMarkdown                 int
		)
		if err := rows.Scan(&id, &dateVal, &title, &content, &weather, &temp, &tempOut, &isPublic, &isMarkdown); err != nil {
			return nil, err
		}
		title = util.UnicodeDecode(title)
		if title != "账单" {
			continue
		}
		dateStr := dateVal.Format("2006-01-02 15:04:05")
		return map[string]interface{}{
			"id":                  id,
			"date":                dateStr,
			"title":               title,
			"content":             util.UnicodeDecode(content),
			"weather":             weather,
			"temperature":         temp,
			"temperature_outside": tempOut,
			"is_public":           isPublic,
			"is_markdown":         isMarkdown,
		}, nil
	}
	return nil, rows.Err()
}

func saveBillDiary(uid int64, date string, content string, current map[string]interface{}) error {
	diary, err := openDB()
	if err != nil {
		return err
	}
	now := util.NowString()
	if current != nil {
		id := current["id"].(int64)
		dateVal, _ := current["date"].(string)
		weather, _ := current["weather"].(string)
		if weather == "" {
			weather = "sunny"
		}
		temp, _ := current["temperature"].(float64)
		tempOut, _ := current["temperature_outside"].(float64)
		isPublic, _ := current["is_public"].(int)
		isMarkdown, _ := current["is_markdown"].(int)
		_, err = diary.Exec(`
			UPDATE diaries SET date_modify=?, date=?, category=?, title=?, content=?, weather=?,
				temperature=?, temperature_outside=?, is_public=?, is_markdown=?
			WHERE id=? AND uid=?`,
			now, dateVal, "bill", util.UnicodeEncode("账单"), util.UnicodeEncode(content), weather,
			temp, tempOut, isPublic, isMarkdown, id, uid)
		return err
	}
	_, err = diary.Exec(`
		INSERT INTO diaries(title, content, category, weather, temperature, temperature_outside,
			date_create, date_modify, date, uid, is_public, is_markdown)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		util.UnicodeEncode("账单"), util.UnicodeEncode(content), "bill", "sunny", -273.0, -273.0,
		now, now, date+" 12:00:00", uid, 0, 0)
	return err
}

// applyLines 合并写入日记；返回写入的合并行、实际落库的日记天数
func applyLines(uid int64, lines []DayLine, known map[string]bool) ([]DayLine, int, error) {
	var fresh []DayLine
	for _, line := range lines {
		var ids []string
		allKnown := len(line.OrderIDs) > 0
		for _, oid := range line.OrderIDs {
			if oid == "" {
				continue
			}
			if known[oid] {
				continue
			}
			ids = append(ids, oid)
			allKnown = false
		}
		if allKnown && len(line.OrderIDs) > 0 {
			continue
		}
		copied := line
		if len(ids) > 0 {
			copied.OrderIDs = ids
		}
		fresh = append(fresh, copied)
	}

	byDate := map[string][]DayLine{}
	var dates []string
	for _, line := range fresh {
		if _, ok := byDate[line.Date]; !ok {
			dates = append(dates, line.Date)
		}
		byDate[line.Date] = append(byDate[line.Date], line)
	}

	var written []DayLine
	diaryCount := 0
	for _, date := range dates {
		current, err := findBillDiary(uid, date)
		if err != nil && err != sql.ErrNoRows {
			return written, diaryCount, err
		}
		var existing []contentRow
		if current != nil {
			existing = parseContent(current["content"].(string))
		}
		pool := []float64{}
		for _, row := range existing {
			if row.Price != nil {
				pool = append(pool, *row.Price)
			}
		}
		var accepted []DayLine
		contentChanged := false
		for _, line := range byDate[date] {
			hit := -1
			for i, price := range pool {
				if math.Abs(price-line.Amount) < 0.011 {
					hit = i
					break
				}
			}
			if hit >= 0 {
				pool = append(pool[:hit], pool[hit+1:]...)
				for _, oid := range line.OrderIDs {
					known[oid] = true
				}
				written = append(written, line) // 记订单号，避免下次再撞
				continue
			}
			merged := false
			for i := range existing {
				if existing[i].Name == line.Name && existing[i].Price != nil {
					*existing[i].Price += line.Amount
					existing[i].Raw = existing[i].Name + " " + fmtAmount(*existing[i].Price)
					merged = true
					contentChanged = true
					break
				}
			}
			if !merged {
				accepted = append(accepted, line)
				contentChanged = true
			}
			for _, oid := range line.OrderIDs {
				known[oid] = true
			}
			written = append(written, line)
		}
		if !contentChanged {
			continue
		}
		var newRows []contentRow
		newRows = append(newRows, existing...)
		for _, line := range accepted {
			newRows = append(newRows, contentRow{
				Name: line.Name,
				Raw:  line.Name + " " + fmtAmount(line.Amount),
			})
		}
		var body []string
		for _, row := range newRows {
			if row.Raw != "" {
				body = append(body, row.Raw)
			}
		}
		content := strings.Join(body, "\n")
		if err := saveBillDiary(uid, date, content, current); err != nil {
			return written, diaryCount, err
		}
		diaryCount++
	}
	return written, diaryCount, nil
}

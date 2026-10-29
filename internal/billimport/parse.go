package billimport

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// decodeCSVBytes 尝试 utf-8-sig / gb18030 / utf-8，保留含「交易时间」的解码结果
func decodeCSVBytes(raw []byte) (string, error) {
	candidates := []func([]byte) (string, error){
		func(b []byte) (string, error) {
			b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
			if !utf8.Valid(b) {
				return "", fmt.Errorf("invalid utf-8")
			}
			return string(b), nil
		},
		func(b []byte) (string, error) {
			r := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GB18030.NewDecoder())
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r); err != nil {
				return "", err
			}
			return buf.String(), nil
		},
	}
	var lastErr error
	for _, fn := range candidates {
		text, err := fn(raw)
		if err != nil {
			lastErr = err
			continue
		}
		if strings.Contains(text, "交易时间") {
			return text, nil
		}
	}
	if lastErr != nil {
		return "", fmt.Errorf("无法识别账单编码: %w", lastErr)
	}
	return "", fmt.Errorf("无法识别账单编码")
}

// parseCSVText 从支付宝导出文本解析交易行
func parseCSVText(text string) ([]RawRow, error) {
	lines := strings.Split(text, "\n")
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimRight(line, "\r"), "交易时间,") {
			start = i
			break
		}
	}
	if start < 0 {
		return nil, fmt.Errorf("未找到交易明细表头")
	}
	body := strings.Join(lines[start:], "\n")
	reader := csv.NewReader(strings.NewReader(body))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV 解析失败: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("没有交易记录")
	}
	header := records[0]
	col := map[string]int{}
	for i, h := range header {
		col[strings.TrimSpace(h)] = i
	}
	need := []string{"交易时间", "交易对方", "商品说明", "交易分类", "收/支", "交易状态", "交易订单号", "金额"}
	for _, k := range need {
		if _, ok := col[k]; !ok {
			return nil, fmt.Errorf("缺少列: %s", k)
		}
	}
	get := func(rec []string, key string) string {
		i := col[key]
		if i >= len(rec) {
			return ""
		}
		return strings.TrimSpace(strings.Trim(rec[i], "\t"))
	}

	var rows []RawRow
	for _, rec := range records[1:] {
		if len(rec) == 0 || strings.TrimSpace(rec[0]) == "" {
			continue
		}
		whenStr := get(rec, "交易时间")
		when, err := time.ParseInLocation("2006-01-02 15:04:05", whenStr, time.Local)
		if err != nil {
			continue
		}
		amount, err := strconv.ParseFloat(get(rec, "金额"), 64)
		if err != nil {
			continue
		}
		rows = append(rows, RawRow{
			When:     when,
			Who:      get(rec, "交易对方"),
			Product:  get(rec, "商品说明"),
			Category: get(rec, "交易分类"),
			IO:       get(rec, "收/支"),
			Status:   get(rec, "交易状态"),
			OrderID:  get(rec, "交易订单号"),
			Amount:   amount,
		})
	}
	return rows, nil
}

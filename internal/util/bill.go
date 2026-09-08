package util

import (
	"regexp"
	"strings"
)

type BillItem struct {
	Item  string  `json:"item"`
	Price float64 `json:"price"`
}

type BillDay struct {
	ID        interface{} `json:"id"`
	MonthID   interface{} `json:"month_id,omitempty"`
	Date      interface{} `json:"date"`
	Items     []BillItem  `json:"items"`
	Sum       float64     `json:"sum"`
	SumIncome float64     `json:"sumIncome"`
	SumOutput float64     `json:"sumOutput"`
}

func ProcessBillOfDay(id, monthID, date interface{}, content string, filterKeywords []string) BillDay {
	str := regexp.MustCompile(` +`).ReplaceAllString(content, " ")
	lines := strings.Split(str, "\n")
	resp := BillDay{ID: id, MonthID: monthID, Date: date, Items: []BillItem{}}
	var re *regexp.Regexp
	if len(filterKeywords) > 0 {
		re = regexp.MustCompile("(?i).*(" + strings.Join(filterKeywords, "|") + ").*")
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if re != nil && !re.MatchString(line) {
			continue
		}
		parts := strings.Split(line, " ")
		price := 0.0
		if len(parts) > 1 {
			price = ParseFloat(parts[1])
		}
		if price < 0 {
			resp.SumOutput = FormatMoney(resp.SumOutput + price)
		} else {
			resp.SumIncome = FormatMoney(resp.SumIncome + price)
		}
		resp.Sum = FormatMoney(resp.Sum + price)
		itemName := parts[0]
		resp.Items = append(resp.Items, BillItem{Item: itemName, Price: price})
	}
	return resp
}

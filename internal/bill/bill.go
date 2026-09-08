package bill

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const dbName = db.Diary

// BillItem 单个账单条目
type BillItem struct {
	Item  string  `json:"item"`
	Price float64 `json:"price"`
}

// BillDay 一天的账单数据
type BillDay struct {
	ID        interface{} `json:"id"`
	MonthID   interface{} `json:"month_id"`
	Date      interface{} `json:"date"`
	Items     []BillItem  `json:"items"`
	Sum       float64     `json:"sum"`
	SumIncome float64     `json:"sumIncome"`
	SumOutput float64     `json:"sumOutput"`
}

// BillKey 账单关键字统计
type BillKey struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

var multiSpaceRe = regexp.MustCompile(` +`)

// ProcessBillOfDay 将日记账单文本内容解析为结构化的账单数据，
// 对应 utility.processBillOfDay。content 应为已解码的日记内容。
func ProcessBillOfDay(row map[string]interface{}, filterKeywords []string) BillDay {
	content := util.UnicodeDecode(asString(row["content"]))
	str := multiSpaceRe.ReplaceAllString(content, " ")
	lines := strings.Split(str, "\n")

	var kept []string
	var kwRe *regexp.Regexp
	if len(filterKeywords) > 0 {
		kwRe = regexp.MustCompile(`(?i).*(` + strings.Join(escapeAll(filterKeywords), "|") + `).*`)
	}
	for _, l := range lines {
		if len(strings.TrimSpace(l)) == 0 {
			continue
		}
		if kwRe != nil && !kwRe.MatchString(l) {
			continue
		}
		kept = append(kept, l)
	}

	resp := BillDay{
		ID:        row["id"],
		MonthID:   row["month_id"],
		Date:      row["date"],
		Items:     []BillItem{},
		Sum:       0,
		SumIncome: 0,
		SumOutput: 0,
	}
	for _, item := range kept {
		infos := strings.Split(item, " ")
		var price float64
		if len(infos) > 1 {
			price, _ = strconv.ParseFloat(infos[1], 64)
		}
		if price < 0 {
			resp.SumOutput = util.FormatMoney(resp.SumOutput + price)
		} else {
			resp.SumIncome = util.FormatMoney(resp.SumIncome + price)
		}
		resp.Sum = util.FormatMoney(resp.Sum + price)
		itemName := ""
		if len(infos) > 0 {
			itemName = infos[0]
		}
		resp.Items = append(resp.Items, BillItem{Item: itemName, Price: price})
	}
	return resp
}

func escapeAll(items []string) []string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = regexp.QuoteMeta(s)
	}
	return out
}

func Register(r *gin.RouterGroup) {
	r.GET("/", handleAll)
	r.GET("/sorted", handleSorted)
	r.GET("/keys", handleKeys)
	r.GET("/day-sum", handleDaySum)
	r.GET("/month-sum", handleMonthSum)
	r.GET("/borrow", handleBorrow)
}

func handleAll(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, `SELECT * from diaries where uid=? and category = 'bill' order by date asc`, user.UID)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	list := []BillDay{}
	for _, row := range rows {
		list = append(list, ProcessBillOfDay(row, nil))
	}
	response.Success(c, list, "请求成功")
}

func monthRows(user int64, year, month int) ([]map[string]interface{}, error) {
	diary, err := db.Open(dbName)
	if err != nil {
		return nil, err
	}
	return db.QueryMaps(diary, `
		select *,
		       date_format(date,'%Y%m') as month_id,
		       date_format(date,'%m') as month
		from diaries
		where year(date) = ? and month(date) = ? and category = 'bill' and uid = ?
		order by date asc`, year, month, user)
}

func splitKeywords(c *gin.Context) []string {
	kw := c.Query("keyword")
	if kw == "" {
		return nil
	}
	return strings.Split(kw, " ")
}

func handleSorted(c *gin.Context) {
	yearsStr := c.Query("years")
	if yearsStr == "" {
		response.Error(c, "", "未选择年份")
		return
	}
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	keywords := splitKeywords(c)

	type monthResp struct {
		ID          interface{} `json:"id"`
		MonthID     interface{} `json:"month_id"`
		Month       interface{} `json:"month"`
		Count       int         `json:"count"`
		Days        []BillDay   `json:"days"`
		Sum         float64     `json:"sum"`
		SumIncome   float64     `json:"sumIncome"`
		SumOutput   float64     `json:"sumOutput"`
		IncomeTop5  []BillItem  `json:"incomeTop5"`
		OutcomeTop5 []BillItem  `json:"outcomeTop5"`
		Food        gin.H       `json:"food"`
	}

	responseData := []monthResp{}
	for _, ys := range strings.Split(yearsStr, ",") {
		year, err := strconv.Atoi(strings.TrimSpace(ys))
		if err != nil {
			continue
		}
		for month := 1; month <= 12; month++ {
			daysArray, err := monthRows(user.UID, year, month)
			if err != nil || len(daysArray) == 0 {
				continue
			}
			var daysData []BillDay
			var monthSum, monthSumIncome, monthSumOutput float64
			var breakfast, launch, dinner, supermarket, fruit float64
			for _, d := range daysArray {
				pd := ProcessBillOfDay(d, keywords)
				if len(pd.Items) > 0 {
					daysData = append(daysData, pd)
					monthSum += pd.Sum
					monthSumIncome += pd.SumIncome
					monthSumOutput += pd.SumOutput
					breakfast += firstNonZero(pd.Items, "早餐")
					launch += firstNonZero(pd.Items, "午餐")
					dinner += firstNonZero(pd.Items, "晚餐")
					supermarket += firstNonZero(pd.Items, "超市")
					fruit += firstNonZero(pd.Items, "水果")
				}
			}
			if len(daysData) == 0 {
				continue
			}
			top5 := getBillMonthTop5(daysData)
			responseData = append(responseData, monthResp{
				ID:          daysArray[0]["id"],
				MonthID:     daysArray[0]["month_id"],
				Month:       daysArray[0]["month"],
				Count:       len(daysArray),
				Days:        daysData,
				Sum:         util.FormatMoney(monthSum),
				SumIncome:   util.FormatMoney(monthSumIncome),
				SumOutput:   util.FormatMoney(monthSumOutput),
				IncomeTop5:  top5.income,
				OutcomeTop5: top5.outcome,
				Food: gin.H{
					"breakfast":   util.FormatMoney(breakfast),
					"launch":      util.FormatMoney(launch),
					"dinner":      util.FormatMoney(dinner),
					"supermarket": util.FormatMoney(supermarket),
					"fruit":       util.FormatMoney(fruit),
					"sum":         util.FormatMoney(breakfast + launch + dinner + supermarket + fruit),
				},
			})
		}
	}
	// reverse
	for i, j := 0, len(responseData)-1; i < j; i, j = i+1, j-1 {
		responseData[i], responseData[j] = responseData[j], responseData[i]
	}
	response.Success(c, responseData, "")
}

func firstNonZero(items []BillItem, keyword string) float64 {
	for _, it := range items {
		if strings.Contains(it.Item, keyword) && it.Price != 0 {
			return it.Price
		}
	}
	return 0
}

type top5Result struct {
	income  []BillItem
	outcome []BillItem
}

func getBillMonthTop5(days []BillDay) top5Result {
	var all []BillItem
	for _, d := range days {
		all = append(all, d.Items...)
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].Price < all[j].Price })

	var income []BillItem
	for _, it := range all {
		if it.Price > 0 {
			income = append(income, it)
		}
	}
	sort.SliceStable(income, func(i, j int) bool { return income[i].Price > income[j].Price })

	outcome := all
	if len(outcome) > 5 {
		outcome = outcome[:5]
	}
	if len(income) > 5 {
		income = income[:5]
	}
	if outcome == nil {
		outcome = []BillItem{}
	}
	if income == nil {
		income = []BillItem{}
	}
	return top5Result{income: income, outcome: outcome}
}

func handleKeys(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	currentYear := time.Now().Year()
	keyMap := map[string]int{}
	order := []string{}
	for i := 0; i < 5; i++ {
		year := currentYear - i
		for month := 1; month <= 12; month++ {
			daysArray, err := monthRows(user.UID, year, month)
			if err != nil {
				continue
			}
			for _, d := range daysArray {
				pd := ProcessBillOfDay(d, nil)
				for _, item := range pd.Items {
					if _, ok := keyMap[item.Item]; !ok {
						order = append(order, item.Item)
					}
					keyMap[item.Item]++
				}
			}
		}
	}
	list := []BillKey{}
	for _, k := range order {
		if keyMap[k] >= 1 {
			list = append(list, BillKey{Key: k, Count: keyMap[k]})
		}
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].Count > list[j].Count })
	response.Success(c, list, "")
}

func handleDaySum(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, `select content, date from diaries where category = 'bill' and uid = ?`, user.UID)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	final := []gin.H{}
	for _, row := range rows {
		pd := ProcessBillOfDay(row, nil)
		final = append(final, gin.H{
			"id":        pd.ID,
			"month_id":  pd.MonthID,
			"date":      pd.Date,
			"sumIncome": pd.SumIncome,
			"sumOutput": pd.SumOutput,
		})
	}
	response.Success(c, final, "获取成功")
}

func handleMonthSum(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	keywords := splitKeywords(c)
	yearNow := time.Now().Year()
	type monthResp struct {
		ID        interface{} `json:"id"`
		MonthID   interface{} `json:"month_id"`
		Month     interface{} `json:"month"`
		Count     int         `json:"count"`
		Sum       float64     `json:"sum"`
		SumIncome float64     `json:"sumIncome"`
		SumOutput float64     `json:"sumOutput"`
	}
	responseData := []monthResp{}
	for year := 2018; year <= yearNow; year++ {
		for month := 1; month <= 12; month++ {
			daysArray, err := monthRows(user.UID, year, month)
			if err != nil || len(daysArray) == 0 {
				continue
			}
			var daysData int
			var monthSum, monthSumIncome, monthSumOutput float64
			for _, d := range daysArray {
				pd := ProcessBillOfDay(d, keywords)
				if len(pd.Items) > 0 {
					daysData++
					monthSum += pd.Sum
					monthSumIncome += pd.SumIncome
					monthSumOutput += pd.SumOutput
				}
			}
			if daysData > 0 {
				responseData = append(responseData, monthResp{
					ID:        daysArray[0]["id"],
					MonthID:   daysArray[0]["month_id"],
					Month:     daysArray[0]["month"],
					Count:     len(daysArray),
					Sum:       util.FormatMoney(monthSum),
					SumIncome: util.FormatMoney(monthSumIncome),
					SumOutput: util.FormatMoney(monthSumOutput),
				})
			}
		}
	}
	response.Success(c, responseData, "")
}

func handleBorrow(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	diary, _ := db.Open(dbName)
	row, err := db.QueryMap(diary, `select * from diaries where title = '借还记录' and uid = ?`, user.UID)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	if row == nil {
		response.Success(c, "", "")
		return
	}
	content := util.UnicodeDecode(asString(row["content"]))
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, content, "")
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

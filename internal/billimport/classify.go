package billimport

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

var skipStatus = map[string]bool{
	"交易关闭":     true,
	"退款成功":     true,
	"等待对方确认收货": true,
}

var skipIO = map[string]bool{
	"不计收支": true,
}

var okStatus = map[string]bool{
	"交易成功":   true,
	"等待确认收货": true,
	"支付成功":   true,
}

func mealSlot(slots []Slot, when time.Time) string {
	hour := float64(when.Hour()) + float64(when.Minute())/60
	for _, slot := range slots {
		if slot.Start <= hour && hour < slot.End {
			return slot.Name
		}
	}
	return "餐饮"
}

func matchRule(rules []Rule, who, product, category string) *Rule {
	text := who + " " + product + " " + category
	for i := range rules {
		rule := &rules[i]
		if len(rule.WhenProduct) > 0 {
			hit := false
			for _, word := range rule.WhenProduct {
				if word != "" && strings.Contains(product, word) {
					hit = true
					break
				}
			}
			if !hit {
				continue
			}
		}
		for _, word := range rule.Match {
			if word != "" && strings.Contains(text, word) {
				return rule
			}
		}
	}
	return nil
}

func shortName(rule *Rule, slots []Slot, when time.Time) string {
	if rule.MealPlace == "" {
		return rule.Name
	}
	slot := mealSlot(slots, when)
	if rule.SlotNames != nil {
		if mapped, ok := rule.SlotNames[slot]; ok {
			slot = mapped
		}
	}
	if rule.MealOrder == "place_first" {
		return rule.MealPlace + "-" + slot
	}
	return slot + "-" + rule.MealPlace
}

func fmtAmount(value float64) string {
	abs := math.Abs(value)
	text := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", abs), "0"), ".")
	if text == "" {
		text = "0"
	}
	if value > 0 {
		return "+" + text
	}
	if value < 0 {
		return "-" + text
	}
	return text
}

func ruleMatchKey(rule Rule) string {
	return strings.Join(rule.Match, ",")
}

// ruleHitsRow 单条规则是否命中（不看规则顺序抢占）
func ruleHitsRow(rule Rule, who, product, category string) bool {
	if len(rule.Match) == 0 {
		return false
	}
	if len(rule.WhenProduct) > 0 {
		hit := false
		for _, word := range rule.WhenProduct {
			if word != "" && strings.Contains(product, word) {
				hit = true
				break
			}
		}
		if !hit {
			return false
		}
	}
	text := who + " " + product + " " + category
	for _, word := range rule.Match {
		if word != "" && strings.Contains(text, word) {
			return true
		}
	}
	return false
}

func signedAmount(row RawRow) float64 {
	if row.IO == "收入" {
		return row.Amount
	}
	return -row.Amount
}

func floatPtr(v float64) *float64 {
	return &v
}

func setRuleRefFromRow(rule *Rule, row RawRow) {
	amt := signedAmount(row)
	rule.RefWhen = row.When.Format("2006-01-02 15:04")
	rule.RefAmount = floatPtr(amt)
	rule.RefWho = row.Who
	rule.RefProduct = row.Product
}

func setRuleRefFromItem(rule *Rule, item Item) {
	amt := item.Amount
	rule.RefWhen = item.When.Format("2006-01-02 15:04")
	rule.RefAmount = floatPtr(amt)
	rule.RefWho = item.Who
	rule.RefProduct = item.Product
}

// buildRuleRefs 为每条别名规则找当前批次样例（最近一笔 + 命中次数）
func buildRuleRefs(rows []RawRow, cfg AliasConfig) []RuleRef {
	refs := make([]RuleRef, 0, len(cfg.Rules))
	for _, rule := range cfg.Rules {
		var best *RawRow
		count := 0
		for i := range rows {
			row := &rows[i]
			if !ruleHitsRow(rule, row.Who, row.Product, row.Category) {
				continue
			}
			count++
			if best == nil || row.When.After(best.When) {
				best = row
			}
		}
		ref := RuleRef{MatchKey: ruleMatchKey(rule), Count: count}
		if best != nil {
			ref.When = best.When.Format("2006-01-02 15:04")
			ref.Amount = signedAmount(*best)
			ref.Who = best.Who
			ref.Product = best.Product
		} else if rule.RefWhen != "" {
			// 本批次未命中时回退已持久化的参考值
			ref.When = rule.RefWhen
			if rule.RefAmount != nil {
				ref.Amount = *rule.RefAmount
			}
			ref.Who = rule.RefWho
			ref.Product = rule.RefProduct
		}
		refs = append(refs, ref)
	}
	return refs
}

// persistRuleRefs 用当前批次样例写回规则参考值；有更新返回 true
func persistRuleRefs(cfg AliasConfig, rows []RawRow) (AliasConfig, bool) {
	changed := false
	for i := range cfg.Rules {
		rule := &cfg.Rules[i]
		var best *RawRow
		for j := range rows {
			row := &rows[j]
			if !ruleHitsRow(*rule, row.Who, row.Product, row.Category) {
				continue
			}
			if best == nil || row.When.After(best.When) {
				best = row
			}
		}
		if best == nil {
			continue
		}
		amt := signedAmount(*best)
		when := best.When.Format("2006-01-02 15:04")
		if rule.RefWhen == when && rule.RefWho == best.Who && rule.RefProduct == best.Product &&
			rule.RefAmount != nil && *rule.RefAmount == amt {
			continue
		}
		setRuleRefFromRow(rule, *best)
		changed = true
	}
	return cfg, changed
}

// classify 将原始行分到 ready/guess/confirm/unmapped/skip
// 金额为 0 不再一律跳过，便于预览辨认；写入时默认仍可过滤
func classify(rows []RawRow, cfg AliasConfig) []Item {
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		signed := -row.Amount
		if row.IO == "收入" {
			signed = row.Amount
		}
		item := Item{
			When:    row.When,
			Date:    row.When.Format("2006-01-02"),
			Who:     row.Who,
			Product: row.Product,
			IO:      row.IO,
			Status:  row.Status,
			OrderID: row.OrderID,
			Amount:  signed,
			Bucket:  bucketUnmapped,
		}
		if skipIO[row.IO] || skipStatus[row.Status] {
			item.Bucket = bucketSkip
			if row.Status != "" {
				item.Reason = row.Status
			} else {
				item.Reason = row.IO
			}
			items = append(items, item)
			continue
		}
		if !okStatus[row.Status] {
			item.Bucket = bucketSkip
			if row.Status != "" {
				item.Reason = row.Status
			} else {
				item.Reason = "状态不明"
			}
			items = append(items, item)
			continue
		}
		rule := matchRule(cfg.Rules, row.Who, row.Product, row.Category)
		if rule == nil {
			items = append(items, item)
			continue
		}
		item.Name = shortName(rule, cfg.Slots, row.When)
		switch {
		case rule.Confirm:
			item.Bucket = bucketConfirm
		case rule.Guess:
			item.Bucket = bucketGuess
		default:
			item.Bucket = bucketReady
		}
		items = append(items, item)
	}
	return items
}

// markImported 已在台账中的订单号标成「已导入」，不再进可写入等桶
func markImported(items []Item, known map[string]bool) {
	if len(known) == 0 {
		return
	}
	for i := range items {
		oid := items[i].OrderID
		if oid == "" || !known[oid] {
			continue
		}
		items[i].Bucket = bucketImported
		items[i].Reason = "已导入"
		if items[i].Name == "" {
			if items[i].Who != "" {
				items[i].Name = items[i].Who
			} else {
				items[i].Name = "已导入"
			}
		}
	}
}

func countBuckets(items []Item) Counts {
	c := Counts{Row: len(items)}
	for _, item := range items {
		switch item.Bucket {
		case bucketReady:
			c.Ready++
		case bucketGuess:
			c.Guess++
		case bucketConfirm:
			c.Confirm++
		case bucketImported:
			c.Imported++
		case bucketUnmapped:
			c.Unmapped++
		case bucketSkip:
			c.Skip++
		}
	}
	return c
}

// groupLines 同一天、同一短名称合并金额
func groupLines(items []Item) []DayLine {
	type key struct{ date, name string }
	grouped := map[key]*DayLine{}
	order := []key{}
	for _, item := range items {
		k := key{item.Date, item.Name}
		if g, ok := grouped[k]; ok {
			g.Amount += item.Amount
			g.OrderIDs = append(g.OrderIDs, item.OrderID)
			if item.When.Before(g.When) {
				g.When = item.When
			}
			continue
		}
		grouped[k] = &DayLine{
			Date:     item.Date,
			Name:     item.Name,
			Amount:   item.Amount,
			When:     item.When,
			OrderIDs: []string{item.OrderID},
			Bucket:   item.Bucket,
		}
		order = append(order, k)
	}
	lines := make([]DayLine, 0, len(order))
	for _, k := range order {
		lines = append(lines, *grouped[k])
	}
	sort.Slice(lines, func(i, j int) bool {
		if lines[i].Date != lines[j].Date {
			return lines[i].Date < lines[j].Date
		}
		if !lines[i].When.Equal(lines[j].When) {
			return lines[i].When.Before(lines[j].When)
		}
		return lines[i].Name < lines[j].Name
	})
	return lines
}

func renderDays(lines []DayLine) string {
	byDate := map[string][]DayLine{}
	var dates []string
	for _, line := range lines {
		if _, ok := byDate[line.Date]; !ok {
			dates = append(dates, line.Date)
		}
		byDate[line.Date] = append(byDate[line.Date], line)
	}
	sort.Strings(dates)
	var parts []string
	for _, date := range dates {
		var body []string
		for _, line := range byDate[date] {
			body = append(body, line.Name+" "+fmtAmount(line.Amount))
		}
		parts = append(parts, date+"\n"+strings.Join(body, "\n"))
	}
	return strings.Join(parts, "\n\n")
}

func buildPreview(items []Item) BucketPreview {
	return buildPreviewWithRows(items, nil, AliasConfig{})
}

func buildPreviewWithRows(items []Item, rows []RawRow, cfg AliasConfig) BucketPreview {
	var readyItems, guessItems, confirmItems, importedItems, unmapped, skip []Item
	for _, item := range items {
		switch item.Bucket {
		case bucketReady:
			readyItems = append(readyItems, item)
		case bucketGuess:
			guessItems = append(guessItems, item)
		case bucketConfirm:
			confirmItems = append(confirmItems, item)
		case bucketImported:
			importedItems = append(importedItems, item)
		case bucketUnmapped:
			unmapped = append(unmapped, item)
		case bucketSkip:
			skip = append(skip, item)
		}
	}
	ready := groupLines(readyItems)
	guess := groupLines(guessItems)
	confirm := groupLines(confirmItems)
	imported := groupLines(importedItems)
	preview := BucketPreview{
		Ready:        ready,
		Guess:        guess,
		Confirm:      confirm,
		Imported:     imported,
		Unmapped:     unmapped,
		Skip:         skip,
		ReadyDays:    renderDays(ready),
		GuessDays:    renderDays(guess),
		ConfirmDays:  renderDays(confirm),
		ImportedDays: renderDays(imported),
		Counts:       countBuckets(items),
	}
	if rows != nil {
		preview.RuleRefs = buildRuleRefs(rows, cfg)
	}
	return preview
}

// seedRulesFromUnmapped 为未识别对方追加 guess 草稿，不覆盖已有 match 词
func seedRulesFromUnmapped(cfg AliasConfig, unmapped []Item) (AliasConfig, int) {
	existing := map[string]bool{}
	for _, rule := range cfg.Rules {
		for _, m := range rule.Match {
			if m != "" {
				existing[m] = true
			}
		}
	}
	added := 0
	seen := map[string]bool{}
	for _, item := range unmapped {
		key := strings.TrimSpace(item.Who)
		useProduct := key == "" || strings.Contains(key, "*")
		if useProduct {
			key = strings.TrimSpace(item.Product)
		}
		if key == "" || seen[key] || existing[key] {
			continue
		}
		// 截断过长商品说明
		matchKey := key
		if len([]rune(matchKey)) > 40 {
			matchKey = string([]rune(matchKey)[:40])
		}
		display := matchKey
		if len([]rune(display)) > 24 {
			display = string([]rune(display)[:24])
		}
		rule := Rule{
			Match: []string{matchKey},
			Name:  "待命名-" + display,
			Guess: true,
		}
		setRuleRefFromItem(&rule, item)
		cfg.Rules = append(cfg.Rules, rule)
		seen[key] = true
		existing[matchKey] = true
		added++
	}
	return cfg, added
}

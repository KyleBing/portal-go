package billimport

import "time"

// Slot 餐次时段（小时含分钟，10:30 = 10.5）
type Slot struct {
	Name  string  `json:"name"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// Rule 别名规则，按数组顺序匹配
type Rule struct {
	Match       []string          `json:"match"`
	Name        string            `json:"name,omitempty"`
	MealPlace   string            `json:"meal_place,omitempty"`
	WhenProduct []string          `json:"when_product,omitempty"`
	SlotNames   map[string]string `json:"slot_names,omitempty"`
	MealOrder   string            `json:"meal_order,omitempty"` // place_first 拼成「盒马-午餐」
	Guess       bool              `json:"guess,omitempty"`
	Confirm     bool              `json:"confirm,omitempty"`
	// 参考流水：来自导入批次样例，便于辨认规则对应哪笔交易
	RefWhen    string   `json:"ref_when,omitempty"`
	RefAmount  *float64 `json:"ref_amount,omitempty"`
	RefWho     string   `json:"ref_who,omitempty"`
	RefProduct string   `json:"ref_product,omitempty"`
}

// AliasConfig 用户别名配置
type AliasConfig struct {
	Slots []Slot `json:"slots"`
	Rules []Rule `json:"rules"`
}

// RawRow 支付宝 CSV 一行
type RawRow struct {
	When     time.Time
	Who      string
	Product  string
	Category string
	IO       string
	Status   string
	OrderID  string
	Amount   float64
}

// Item 分类后的单笔
type Item struct {
	When    time.Time `json:"when"`
	Date    string    `json:"date"`
	Who     string    `json:"who"`
	Product string    `json:"product"`
	IO      string    `json:"io"`
	Status  string    `json:"status"`
	OrderID string    `json:"order_id"`
	Amount  float64   `json:"amount"`
	Name    string    `json:"name"`
	Bucket  string    `json:"bucket"`
	Reason  string    `json:"reason,omitempty"`
}

// DayLine 同一天同一短名称合并后的行
type DayLine struct {
	Date     string    `json:"date"`
	Name     string    `json:"name"`
	Amount   float64   `json:"amount"`
	When     time.Time `json:"when"`
	OrderIDs []string  `json:"order_ids"`
	Bucket   string    `json:"bucket"`
}

// RuleRef 别名规则在当前批次中的样例流水，方便辨认是哪家店
type RuleRef struct {
	MatchKey string  `json:"match_key"`
	When     string  `json:"when"`
	Amount   float64 `json:"amount"`
	Who      string  `json:"who"`
	Product  string  `json:"product"`
	Count    int     `json:"count"`
}

// BucketPreview 预览分桶结果
type BucketPreview struct {
	Ready        []DayLine `json:"ready"`
	Guess        []DayLine `json:"guess"`
	Confirm      []DayLine `json:"confirm"`
	Imported     []DayLine `json:"imported"`
	Unmapped     []Item    `json:"unmapped"`
	Skip         []Item    `json:"skip"`
	ReadyDays    string    `json:"ready_days,omitempty"`
	GuessDays    string    `json:"guess_days,omitempty"`
	ConfirmDays  string    `json:"confirm_days,omitempty"`
	ImportedDays string    `json:"imported_days,omitempty"`
	Counts       Counts    `json:"counts"`
	RuleRefs     []RuleRef `json:"rule_refs,omitempty"`
}

// Counts 各桶数量（按原始行）
type Counts struct {
	Row      int `json:"row"`
	Ready    int `json:"ready"`
	Guess    int `json:"guess"`
	Confirm  int `json:"confirm"`
	Imported int `json:"imported"`
	Unmapped int `json:"unmapped"`
	Skip     int `json:"skip"`
}

const (
	bucketReady    = "ready"
	bucketGuess    = "guess"
	bucketConfirm  = "confirm"
	bucketImported = "imported"
	bucketUnmapped = "unmapped"
	bucketSkip     = "skip"

	statusUploaded = "uploaded"
	statusApplied  = "applied"
	statusPartial  = "partial"
)

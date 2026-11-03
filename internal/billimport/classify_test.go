package billimport

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// sampleAliasConfig 测试用样例规则，不作为新用户默认配置
func sampleAliasConfig(t *testing.T) AliasConfig {
	t.Helper()
	raw, err := os.ReadFile("default_aliases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cfg AliasConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestEmptyAliasConfigHasNoRules(t *testing.T) {
	cfg := emptyAliasConfig()
	if len(cfg.Slots) == 0 {
		t.Fatal("slots should keep generic meal times")
	}
	if len(cfg.Rules) != 0 {
		t.Fatalf("new user must not inherit sample rules, got %d", len(cfg.Rules))
	}
}

func TestSampleAliasesLoad(t *testing.T) {
	cfg := sampleAliasConfig(t)
	if len(cfg.Slots) == 0 || len(cfg.Rules) == 0 {
		t.Fatalf("empty sample: slots=%d rules=%d", len(cfg.Slots), len(cfg.Rules))
	}
}

func TestClassifyChaoYiXingBreakfast(t *testing.T) {
	cfg := sampleAliasConfig(t)
	when := time.Date(2026, 10, 26, 6, 44, 0, 0, time.Local)
	items := classify([]RawRow{{
		When: when, Who: "济南和仟运营管理有限公司", Product: "金舆店-店内购物",
		Category: "餐饮美食", IO: "支出", Status: "交易成功", OrderID: "1", Amount: 11.2,
	}}, cfg)
	if len(items) != 1 || items[0].Bucket != bucketReady {
		t.Fatalf("got %+v", items)
	}
	if items[0].Name != "早午餐-超意兴" {
		t.Fatalf("name=%q", items[0].Name)
	}
}

func TestParseSampleCSV(t *testing.T) {
	path := filepath.Join("..", "..", "..", "github", "alipay-diary-bill", "bills")
	// 优先用仓库旁的样例；没有就跳过
	candidates := []string{
		"/Users/kyle/github/alipay-diary-bill/bills/支付宝交易明细(20260917-20261029).csv",
		"/Users/kyle/Downloads/支付宝交易明细(20260926-20261026).csv",
	}
	var raw []byte
	for _, p := range candidates {
		b, err := os.ReadFile(p)
		if err == nil {
			raw = b
			break
		}
	}
	if raw == nil {
		t.Skip("no sample csv")
	}
	_ = path
	text, err := decodeCSVBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := parseCSVText(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("no rows")
	}
	cfg := sampleAliasConfig(t)
	items := classify(rows, cfg)
	c := countBuckets(items)
	if c.Row != len(rows) {
		t.Fatalf("row mismatch %d vs %d", c.Row, len(rows))
	}
	t.Logf("counts: %+v", c)
}

package billimport

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/setup"
	"github.com/gin-gonic/gin"
)

func handleGetAliases(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	cfg, err := ensureAliasConfig(user.UID)
	if err != nil {
		response.Error(c, err.Error(), "读取别名失败")
		return
	}
	response.Success(c, cfg, "")
}

func handlePutAliases(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var cfg AliasConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.Error(c, err.Error(), "参数错误")
		return
	}
	if err := saveAliasConfig(user.UID, cfg); err != nil {
		response.Error(c, err.Error(), "保存别名失败")
		return
	}
	response.Success(c, cfg, "已保存")
}

func handleUpload(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, err.Error(), "未找到上传文件")
		return
	}
	name := filepath.Base(fileHeader.Filename)
	if name == "." || name == ".." || name == "" || strings.Contains(name, "\x00") {
		response.Error(c, "", "非法文件名")
		return
	}
	if !strings.HasSuffix(strings.ToLower(name), ".csv") {
		response.Error(c, "", "请上传 CSV 文件")
		return
	}

	cfg, err := ensureAliasConfig(user.UID)
	if err != nil {
		response.Error(c, err.Error(), "读取别名失败")
		return
	}

	// 先落到临时路径解析，再建 batch 后改名
	dirRel := filepath.Join("upload", strconv.FormatInt(user.UID, 10), "bill-import")
	dirAbs := filepath.Join(setup.ProjectRoot(), dirRel)
	if err := os.MkdirAll(dirAbs, 0o755); err != nil {
		response.Error(c, err.Error(), "创建目录失败")
		return
	}
	tmpName := fmt.Sprintf("tmp_%d_%s", user.UID, name)
	tmpAbs := filepath.Join(dirAbs, tmpName)
	if err := c.SaveUploadedFile(fileHeader, tmpAbs); err != nil {
		response.Error(c, err.Error(), "上传失败")
		return
	}
	raw, err := os.ReadFile(tmpAbs)
	if err != nil {
		_ = os.Remove(tmpAbs)
		response.Error(c, err.Error(), "读取文件失败")
		return
	}
	text, err := decodeCSVBytes(raw)
	if err != nil {
		_ = os.Remove(tmpAbs)
		response.Error(c, err.Error(), err.Error())
		return
	}
	rows, err := parseCSVText(text)
	if err != nil {
		_ = os.Remove(tmpAbs)
		response.Error(c, err.Error(), err.Error())
		return
	}
	// 用本批次样例回写别名参考值
	cfg, refsChanged := persistRuleRefs(cfg, rows)
	if refsChanged {
		_ = saveAliasConfig(user.UID, cfg)
	}
	items := classify(rows, cfg)
	if known, err := knownOrderIDs(user.UID); err == nil {
		markImported(items, known)
	}
	preview := buildPreviewWithRows(items, rows, cfg)
	ps, pe := periodFromRows(rows)

	// 先插入 batch 拿 id，再移动文件
	batchID, err := insertBatch(user.UID, name, "", ps, pe, preview.Counts)
	if err != nil {
		_ = os.Remove(tmpAbs)
		response.Error(c, err.Error(), "创建导入批次失败")
		return
	}
	finalName := fmt.Sprintf("%d_%s", batchID, name)
	finalRel := filepath.ToSlash(filepath.Join(dirRel, finalName))
	finalAbs := filepath.Join(dirAbs, finalName)
	if err := os.Rename(tmpAbs, finalAbs); err != nil {
		_ = os.Remove(tmpAbs)
		response.Error(c, err.Error(), "保存文件失败")
		return
	}
	diary, _ := openDB()
	_, _ = diary.Exec(`UPDATE bill_import_batch SET stored_path=? WHERE id=? AND uid=?`, finalRel, batchID, user.UID)

	batch, _ := getBatch(user.UID, batchID)
	data := gin.H{
		"batch_id": batchID,
		"preview":  preview,
		"aliases":  cfg,
	}
	if batch != nil {
		data["batch"] = batchToMap(batch)
		data["batch"].(map[string]interface{})["stored_path"] = finalRel
	}
	response.Success(c, data, "上传成功")
}

func handlePreview(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	batchID, err := strconv.ParseInt(c.Query("batchId"), 10, 64)
	if err != nil || batchID <= 0 {
		response.Error(c, "", "batchId 无效")
		return
	}
	batch, err := getBatch(user.UID, batchID)
	if err != nil {
		response.Error(c, err.Error(), "批次不存在")
		return
	}
	cfg, err := ensureAliasConfig(user.UID)
	if err != nil {
		response.Error(c, err.Error(), "读取别名失败")
		return
	}
	rows, err := loadBatchRows(batch.StoredPath)
	if err != nil {
		response.Error(c, err.Error(), "读取账单文件失败")
		return
	}
	cfg, refsChanged := persistRuleRefs(cfg, rows)
	if refsChanged {
		_ = saveAliasConfig(user.UID, cfg)
	}
	items := classify(rows, cfg)
	if known, err := knownOrderIDs(user.UID); err == nil {
		markImported(items, known)
	}
	preview := buildPreviewWithRows(items, rows, cfg)
	_ = updateBatchCounts(batchID, preview.Counts)
	response.Success(c, gin.H{
		"batch_id": batchID,
		"preview":  preview,
		"batch":    batchToMap(batch),
		"aliases":  cfg,
	}, "")
}

func handleApply(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body struct {
		BatchID        int64 `json:"batchId"`
		IncludeGuess   bool  `json:"includeGuess"`
		IncludeConfirm bool  `json:"includeConfirm"`
		IncludeZero    bool  `json:"includeZero"` // 默认不写金额为 0
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.BatchID <= 0 {
		response.Error(c, "", "参数错误")
		return
	}
	batch, err := getBatch(user.UID, body.BatchID)
	if err != nil {
		response.Error(c, err.Error(), "批次不存在")
		return
	}
	cfg, err := ensureAliasConfig(user.UID)
	if err != nil {
		response.Error(c, err.Error(), "读取别名失败")
		return
	}
	rows, err := loadBatchRows(batch.StoredPath)
	if err != nil {
		response.Error(c, err.Error(), "读取账单文件失败")
		return
	}
	known, err := knownOrderIDs(user.UID)
	if err != nil {
		response.Error(c, err.Error(), "读取已导入订单失败")
		return
	}
	items := classify(rows, cfg)
	markImported(items, known)
	buckets := map[string]bool{bucketReady: true}
	if body.IncludeGuess {
		buckets[bucketGuess] = true
	}
	if body.IncludeConfirm {
		buckets[bucketConfirm] = true
	}
	var selected []Item
	for _, item := range items {
		if !buckets[item.Bucket] {
			continue
		}
		// 金额 0 默认可预览，写入需显式勾选
		if !body.IncludeZero && item.Amount == 0 {
			continue
		}
		selected = append(selected, item)
	}
	lines := groupLines(selected)
	if len(lines) == 0 {
		response.Success(c, gin.H{"applied": 0, "diary_count": 0}, "没有可写入的条目")
		return
	}
	written, diaryCount, err := applyLines(user.UID, lines, known)
	if err != nil {
		response.Error(c, err.Error(), "写入日记失败")
		return
	}
	if err := insertOrders(user.UID, body.BatchID, written); err != nil {
		response.Error(c, err.Error(), "记录订单号失败")
		return
	}
	status := statusApplied
	if len(written) < len(lines) {
		status = statusPartial
	}
	_ = updateBatchApply(body.BatchID, len(written), status)
	_ = updateBatchCounts(body.BatchID, countBuckets(items))
	response.Success(c, gin.H{
		"applied":     len(written),
		"diary_count": diaryCount,
		"line_count":  len(written),
		"status":      status,
	}, fmt.Sprintf("写入完成，共 %d 条日记", diaryCount))
}

func handleSeedAliases(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	var body struct {
		BatchID int64 `json:"batchId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.BatchID <= 0 {
		response.Error(c, "", "参数错误")
		return
	}
	batch, err := getBatch(user.UID, body.BatchID)
	if err != nil {
		response.Error(c, err.Error(), "批次不存在")
		return
	}
	cfg, err := ensureAliasConfig(user.UID)
	if err != nil {
		response.Error(c, err.Error(), "读取别名失败")
		return
	}
	rows, err := loadBatchRows(batch.StoredPath)
	if err != nil {
		response.Error(c, err.Error(), "读取账单文件失败")
		return
	}
	items := classify(rows, cfg)
	known, _ := knownOrderIDs(user.UID)
	markImported(items, known)
	var unmapped []Item
	for _, item := range items {
		if item.Bucket == bucketUnmapped {
			unmapped = append(unmapped, item)
		}
	}
	cfg, added := seedRulesFromUnmapped(cfg, unmapped)
	cfg, refsChanged := persistRuleRefs(cfg, rows)
	if added > 0 || refsChanged {
		if err := saveAliasConfig(user.UID, cfg); err != nil {
			response.Error(c, err.Error(), "保存别名失败")
			return
		}
	}
	// 重新预览
	items = classify(rows, cfg)
	markImported(items, known)
	preview := buildPreviewWithRows(items, rows, cfg)
	_ = updateBatchCounts(body.BatchID, preview.Counts)
	response.Success(c, gin.H{
		"added":   added,
		"aliases": cfg,
		"preview": preview,
	}, fmt.Sprintf("已追加 %d 条草稿别名", added))
}

func handleListBatches(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	pageNo, _ := strconv.Atoi(c.DefaultQuery("pageNo", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := listBatches(user.UID, pageNo, pageSize)
	if err != nil {
		response.Error(c, err.Error(), "读取导入记录失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total, "pageNo": pageNo, "pageSize": pageSize}, "")
}

func handleListOrders(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", msg)
		return
	}
	batchID, err := strconv.ParseInt(c.Query("batchId"), 10, 64)
	if err != nil || batchID <= 0 {
		response.Error(c, "", "batchId 无效")
		return
	}
	if _, err := getBatch(user.UID, batchID); err != nil {
		response.Error(c, err.Error(), "批次不存在")
		return
	}
	list, err := listOrders(user.UID, batchID)
	if err != nil {
		response.Error(c, err.Error(), "读取订单失败")
		return
	}
	response.Success(c, list, "")
}

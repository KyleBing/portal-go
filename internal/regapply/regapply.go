package regapply

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"log"
	"strings"

	"github.com/KyleBing/portal-go/internal/apihelper"
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/mail"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/systemconfig"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const table = "registration_applications"
const inviteTable = "invitations"

// Register 挂载 /registration-application 路由。
func Register(r *gin.RouterGroup) {
	g := r
	g.POST("/submit", handleSubmit)
	g.GET("/manage", handleManage)
	g.GET("/detail", handleDetail)
	g.POST("/approve", handleApprove)
	g.POST("/reject", handleReject)
	g.DELETE("/delete", handleDelete)
}

func diaryDB() (*sql.DB, error) { return db.Open(db.Diary) }

func handleSubmit(c *gin.Context) {
	body := apihelper.Body(c)
	email := strings.TrimSpace(strings.ToLower(apihelper.S(body, "email")))
	paidAccount := strings.TrimSpace(apihelper.S(body, "paidAccount"))
	if paidAccount == "" {
		paidAccount = strings.TrimSpace(apihelper.S(body, "paid_account"))
	}
	platform := strings.TrimSpace(strings.ToLower(apihelper.S(body, "payPlatform")))
	if platform == "" {
		platform = strings.TrimSpace(strings.ToLower(apihelper.S(body, "pay_platform")))
	}

	if email == "" || !strings.Contains(email, "@") {
		response.Error(c, "", "请填写有效邮箱")
		return
	}
	if paidAccount == "" {
		response.Error(c, "", "请填写付费昵称")
		return
	}
	if platform != "alipay" && platform != "wechat" {
		response.Error(c, "", "请选择付费平台：支付宝或微信")
		return
	}

	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "提交失败")
		return
	}
	var pending int
	if err := diary.QueryRow(`SELECT COUNT(1) FROM `+table+` WHERE email = ? AND status = 'pending'`, email).Scan(&pending); err != nil {
		response.Error(c, err.Error(), "提交失败")
		return
	}
	if pending > 0 {
		response.Error(c, "", "该邮箱已有待审核申请，请耐心等待邮件通知")
		return
	}

	cfg, _ := systemconfig.GetAdminSystemConfig()
	payAmountYuan := cfg.RegistrationPayAmountYuan

	now := util.NowString()
	res, err := diary.Exec(`INSERT INTO `+table+` (email, paid_account, pay_platform, pay_amount_yuan, status, date_create)
		VALUES (?, ?, ?, ?, 'pending', ?)`, email, paidAccount, platform, payAmountYuan, now)
	if err != nil {
		response.Error(c, err.Error(), "提交失败")
		return
	}
	id, _ := res.LastInsertId()
	if err := mail.Default().SendRegistrationApplicationSubmittedEmail(email); err != nil {
		log.Printf("[regapply] submit confirmation mail to %s: %v", email, err)
	}
	response.Success(c, gin.H{"id": id}, "申请已提交，审核结果将发送至您的邮箱")
}

func handleManage(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, errMsg, errMsg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权限操作")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}

	status := c.Query("status") // all | pending | approved | rejected
	pageNo := util.AtoiDefault(c.Query("pageNo"), 1)
	pageSize := util.AtoiDefault(c.Query("pageSize"), 20)
	if pageSize < 1 {
		pageSize = 20
	}
	start := apihelper.PageStart(pageNo, pageSize)

	where := ""
	var countArgs []interface{}
	switch status {
	case "pending", "approved", "rejected":
		where = ` WHERE ra.status = ?`
		countArgs = append(countArgs, status)
	}

	var filteredTotal int64
	if err := diary.QueryRow(`SELECT COUNT(*) FROM `+table+` ra`+where, countArgs...).Scan(&filteredTotal); err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}

	sqlBuf := `SELECT ra.id, ra.email, ra.paid_account, ra.pay_platform, ra.pay_amount_yuan, ra.status, ra.invitation_code, ra.admin_note,
		ra.date_create, ra.date_review, ra.reviewer_uid,
		CASE WHEN ra.invitation_code IS NOT NULL AND ra.invitation_code != '' AND inv.id IS NULL THEN 1 ELSE 0 END AS invitation_missing
		FROM ` + table + ` ra
		LEFT JOIN ` + inviteTable + ` inv ON inv.id = ra.invitation_code` + where + `
		ORDER BY ra.date_create DESC LIMIT ?, ?`
	listArgs := append([]interface{}{}, countArgs...)
	listArgs = append(listArgs, start, pageSize)
	list, err := apihelper.QueryMaps(diary, sqlBuf, listArgs...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}

	var total, pending, approved, rejected int64
	_ = diary.QueryRow(`
		SELECT COUNT(*),
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END), 0)
		FROM `+table).Scan(&total, &pending, &approved, &rejected)

	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{
		"list": list,
		"pager": gin.H{
			"pageNo":   pageNo,
			"pageSize": pageSize,
			"total":    filteredTotal,
		},
		"summary": gin.H{
			"total":    total,
			"pending":  pending,
			"approved": approved,
			"rejected": rejected,
		},
	}, "请求成功")
}

func handleDetail(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, errMsg, errMsg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权限操作")
		return
	}
	id := util.AtoiDefault(c.Query("id"), 0)
	if id <= 0 {
		response.Error(c, "", "参数错误，缺少 id")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	row, err := apihelper.QueryMap(diary, `SELECT ra.id, ra.email, ra.paid_account, ra.pay_platform, ra.pay_amount_yuan, ra.status, ra.invitation_code, ra.admin_note,
		ra.date_create, ra.date_review, ra.reviewer_uid,
		CASE WHEN ra.invitation_code IS NOT NULL AND ra.invitation_code != '' AND inv.id IS NULL THEN 1 ELSE 0 END AS invitation_missing
		FROM `+table+` ra
		LEFT JOIN `+inviteTable+` inv ON inv.id = ra.invitation_code
		WHERE ra.id = ?`, id)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	if row == nil {
		response.Error(c, "", "申请不存在")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, row, "请求成功")
}

func handleApprove(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, errMsg, errMsg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权限操作")
		return
	}
	body := apihelper.Body(c)
	id := util.AtoiDefault(apihelper.S(body, "id"), 0)
	if id <= 0 {
		response.Error(c, "", "参数错误，缺少 id")
		return
	}
	note := strings.TrimSpace(apihelper.S(body, "adminNote"))
	if note == "" {
		note = strings.TrimSpace(apihelper.S(body, "admin_note"))
	}

	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "审核失败")
		return
	}

	var email, paidAccount, status string
	if err := diary.QueryRow(`SELECT email, paid_account, status FROM `+table+` WHERE id = ?`, id).Scan(&email, &paidAccount, &status); err != nil {
		response.Error(c, err.Error(), "审核失败")
		return
	}
	if status != "pending" {
		response.Error(c, "", "该申请已处理，无法重复审核")
		return
	}

	code, err := newInvitationCode(diary, email, paidAccount)
	if err != nil {
		response.Error(c, err.Error(), "生成邀请码失败")
		return
	}

	now := util.NowString()
	_, err = diary.Exec(`UPDATE `+table+` SET status = 'approved', invitation_code = ?, admin_note = ?,
		date_review = ?, reviewer_uid = ? WHERE id = ? AND status = 'pending'`,
		code, nullIfEmpty(note), now, user.UID, id)
	if err != nil {
		response.Error(c, err.Error(), "审核失败")
		return
	}

	if err := mail.Default().SendInvitationEmail(email, code, email, paidAccount); err != nil {
		response.Error(c, err.Error(), "邀请码已生成，但邮件发送失败，请手动通知用户")
		return
	}

	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": id, "invitation_code": code}, "已通过审核，邀请码邮件已发送")
}

func handleReject(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, errMsg, errMsg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权限操作")
		return
	}
	body := apihelper.Body(c)
	id := util.AtoiDefault(apihelper.S(body, "id"), 0)
	if id <= 0 {
		response.Error(c, "", "参数错误，缺少 id")
		return
	}
	note := strings.TrimSpace(apihelper.S(body, "adminNote"))
	if note == "" {
		note = strings.TrimSpace(apihelper.S(body, "admin_note"))
	}

	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "操作失败")
		return
	}
	now := util.NowString()
	res, err := diary.Exec(`UPDATE `+table+` SET status = 'rejected', admin_note = ?,
		date_review = ?, reviewer_uid = ? WHERE id = ? AND status = 'pending'`,
		nullIfEmpty(note), now, user.UID, id)
	if err != nil {
		response.Error(c, err.Error(), "操作失败")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		response.Error(c, "", "该申请已处理或不存在")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": id}, "已拒绝该申请")
}

func handleDelete(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, errMsg, errMsg)
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权限操作")
		return
	}
	id := util.AtoiDefault(c.Query("id"), 0)
	if id <= 0 {
		response.Error(c, "", "参数错误，缺少 id")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "删除失败")
		return
	}
	res, err := diary.Exec(`DELETE FROM `+table+` WHERE id = ?`, id)
	if err != nil {
		response.Error(c, err.Error(), "删除失败")
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		response.Error(c, "", "申请不存在")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": id}, "已删除")
}

// newInvitationCode 审核通过时生成一次性邀请码，并记下申请人与付费昵称。
func newInvitationCode(diary *sql.DB, applicantEmail, paymentUser string) (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	key := base64.StdEncoding.EncodeToString(buf)
	now := util.NowString()
	_, err := diary.Exec(`INSERT INTO `+inviteTable+`(date_create, id, applicant_email, payment_user) VALUES (?, ?, ?, ?)`,
		now, key, nullIfEmpty(applicantEmail), nullIfEmpty(paymentUser))
	return key, err
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

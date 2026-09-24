package user

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/KyleBing/portal-go/internal/apihelper"
	"github.com/KyleBing/portal-go/internal/auth"
	"github.com/KyleBing/portal-go/internal/mail"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/ratelimit"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var authLimits = ratelimit.New()

func issueVerifyToken(diary *sql.DB, uid int64, email string) error {
	raw, hash, err := auth.NewEmailToken()
	if err != nil {
		return err
	}
	expires := time.Now().Add(24 * time.Hour)
	if _, err = diary.Exec(
		`INSERT INTO email_tokens (uid, purpose, token_hash, expires_at) VALUES (?, 'verify', ?, ?)`,
		uid, hash, expires,
	); err != nil {
		return err
	}
	url := mail.Default().AppPublicURL() + "/user/verify?token=" + raw
	return mail.Default().SendVerifyEmail(email, url)
}

func issueResetToken(diary *sql.DB, uid int64, email string) error {
	raw, hash, err := auth.NewEmailToken()
	if err != nil {
		return err
	}
	expires := time.Now().Add(time.Hour)
	if _, err = diary.Exec(
		`INSERT INTO email_tokens (uid, purpose, token_hash, expires_at) VALUES (?, 'reset', ?, ?)`,
		uid, hash, expires,
	); err != nil {
		return err
	}
	url := mail.Default().AppPublicURL() + "/user/reset?token=" + raw
	return mail.Default().SendResetEmail(email, url)
}

// GET /user/verify?token= — 邮件链接验证邮箱，返回 HTML。
func handleVerify(c *gin.Context) {
	raw := strings.TrimSpace(c.Query("token"))
	if raw == "" {
		writeAuthMessage(c, http.StatusBadRequest, "链接无效或已过期", "请重新申请验证邮件。")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		writeAuthMessage(c, http.StatusInternalServerError, "服务异常", "请稍后重试。")
		return
	}
	hash := auth.HashToken(raw)
	var uid int64
	var expires time.Time
	var used sql.NullTime
	err = diary.QueryRow(
		`SELECT uid, expires_at, used_at FROM email_tokens WHERE token_hash = ? AND purpose = 'verify'`,
		hash,
	).Scan(&uid, &expires, &used)
	if err != nil || used.Valid || time.Now().After(expires) {
		writeAuthMessage(c, http.StatusBadRequest, "链接无效或已过期", "请重新申请验证邮件后再试。")
		return
	}
	_, _ = diary.Exec(`UPDATE email_tokens SET used_at = NOW() WHERE token_hash = ?`, hash)
	_, _ = diary.Exec(`UPDATE users SET email_verified_at = NOW() WHERE uid = ? AND email_verified_at IS NULL`, uid)
	writeAuthMessage(c, http.StatusOK, "邮箱已验证", "可以返回日记或管理后台登录了。")
}

// POST /user/resend-verify { email } — 公开重发验证邮件（防枚举：统一成功文案）。
func handleResendVerify(c *gin.Context) {
	body := apihelper.Body(c)
	email := strings.TrimSpace(strings.ToLower(apihelper.S(body, "email")))
	if email == "" {
		response.Error(c, "", "请填写邮箱")
		return
	}
	ip := c.ClientIP()
	if !authLimits.Allow("resend:"+ip, 5, time.Hour) {
		response.Error(c, gin.H{"code": "rate_limited"}, "操作过于频繁，请稍后再试")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "数据库请求出错")
		return
	}
	var uid int64
	var verified sql.NullTime
	err = diary.QueryRow(`SELECT uid, email_verified_at FROM users WHERE email = ?`, email).Scan(&uid, &verified)
	if err == nil && !verified.Valid {
		if authLimits.Allow("resend-uid:"+strconv.FormatInt(uid, 10), 3, time.Hour) {
			_ = issueVerifyToken(diary, uid, email)
		}
	}
	response.Success(c, nil, "若该邮箱已注册且未验证，验证邮件已发送")
}

// POST /user/forgot { email } — 找回密码，始终返回成功。
func handleForgot(c *gin.Context) {
	body := apihelper.Body(c)
	email := strings.TrimSpace(strings.ToLower(apihelper.S(body, "email")))
	ip := c.ClientIP()
	if !authLimits.Allow("forgot:"+ip, 5, time.Hour) {
		response.Success(c, nil, "若该邮箱已注册，重置邮件已发送")
		return
	}
	if email != "" {
		diary, err := diaryDB()
		if err == nil {
			var uid int64
			if err := diary.QueryRow(`SELECT uid FROM users WHERE email = ?`, email).Scan(&uid); err == nil {
				_ = issueResetToken(diary, uid, email)
			}
		}
	}
	response.Success(c, nil, "若该邮箱已注册，重置邮件已发送")
}

// GET /user/reset?token= — 展示重置密码表单；POST { token, password } 提交新密码。
func handleReset(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		raw := strings.TrimSpace(c.Query("token"))
		if raw == "" {
			writeAuthMessage(c, http.StatusBadRequest, "链接无效或已过期", "请重新申请重置密码。")
			return
		}
		diary, err := diaryDB()
		if err != nil {
			writeAuthMessage(c, http.StatusInternalServerError, "服务异常", "请稍后重试。")
			return
		}
		hash := auth.HashToken(raw)
		var uid int64
		var expires time.Time
		var used sql.NullTime
		err = diary.QueryRow(
			`SELECT uid, expires_at, used_at FROM email_tokens WHERE token_hash = ? AND purpose = 'reset'`,
			hash,
		).Scan(&uid, &expires, &used)
		if err != nil || used.Valid || time.Now().After(expires) {
			writeAuthMessage(c, http.StatusBadRequest, "链接无效或已过期", "请重新申请重置密码。")
			return
		}
		writeResetForm(c, raw)
		return
	}

	body := apihelper.Body(c)
	raw := strings.TrimSpace(apihelper.S(body, "token"))
	password := apihelper.S(body, "password")
	if raw == "" || password == "" {
		response.Error(c, "", "参数错误")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "数据库请求出错")
		return
	}
	hash := auth.HashToken(raw)
	var uid int64
	var expires time.Time
	var used sql.NullTime
	err = diary.QueryRow(
		`SELECT uid, expires_at, used_at FROM email_tokens WHERE token_hash = ? AND purpose = 'reset'`,
		hash,
	).Scan(&uid, &expires, &used)
	if err != nil || used.Valid || time.Now().After(expires) {
		response.Error(c, gin.H{"code": "invalid_token"}, "链接无效或已过期")
		return
	}
	pwHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		response.Error(c, "", "重置失败")
		return
	}
	if _, err := diary.Exec(`UPDATE users SET password = ? WHERE uid = ?`, string(pwHash), uid); err != nil {
		response.Error(c, err.Error(), "重置失败")
		return
	}
	_, _ = diary.Exec(`UPDATE email_tokens SET used_at = NOW() WHERE token_hash = ?`, hash)
	// 通过邮件重置密码后视为邮箱已验证
	_, _ = diary.Exec(`UPDATE users SET email_verified_at = COALESCE(email_verified_at, NOW()) WHERE uid = ?`, uid)
	response.Success(c, nil, "密码已更新，可以返回登录")
}

// POST /user/force-verify { uid } — 管理员强制标记邮箱已验证。
func handleForceVerify(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" || user == nil || !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	body := apihelper.Body(c)
	uid := apihelper.I(body, "uid")
	if uid == 0 {
		response.Error(c, "", "参数错误：uid")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "操作失败")
		return
	}
	res, err := diary.Exec(`UPDATE users SET email_verified_at = COALESCE(email_verified_at, NOW()) WHERE uid = ?`, uid)
	if err != nil {
		response.Error(c, err.Error(), "操作失败")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		response.Error(c, "", "用户不存在")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "已标记为已验证")
}

// POST /user/send-reset-password { uid } — 管理员向指定用户发送重置密码邮件。
func handleAdminSendReset(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" || user == nil || !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	body := apihelper.Body(c)
	uid := apihelper.I(body, "uid")
	if uid == 0 {
		response.Error(c, "", "参数错误：uid")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "操作失败")
		return
	}
	var email string
	if err := diary.QueryRow(`SELECT email FROM users WHERE uid = ?`, uid).Scan(&email); err != nil {
		response.Error(c, "", "用户不存在")
		return
	}
	if err := issueResetToken(diary, uid, email); err != nil {
		response.Error(c, err.Error(), "发送重置邮件失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"status": "sent"}, "重置密码邮件已发送")
}

// emailVerifiedFromRow 从 QueryMap 结果判断是否已验证。
func emailVerifiedFromRow(data map[string]interface{}) bool {
	if data == nil {
		return false
	}
	v, ok := data["email_verified_at"]
	if !ok || v == nil {
		return false
	}
	switch t := v.(type) {
	case time.Time:
		return !t.IsZero()
	case *time.Time:
		return t != nil && !t.IsZero()
	case string:
		return strings.TrimSpace(t) != ""
	case []byte:
		return len(t) > 0
	default:
		return true
	}
}

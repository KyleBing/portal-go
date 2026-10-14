package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// 五笔客户端免邀请注册：client 标识 + HMAC 时间窗签名。
// 环境变量 APP_REGISTER_SECRET；未配置则关闭此通道。
const (
	AppClientIOS     = "wubi86-ios"
	AppClientAndroid = "wubi86-android"
	appRegisterSkew  = 5 * time.Minute
)

var allowedAppClients = map[string]struct{}{
	AppClientIOS:     {},
	AppClientAndroid: {},
}

// AppRegisterSecret 读取共享密钥；空表示未启用 App 免邀请注册。
func AppRegisterSecret() string {
	return strings.TrimSpace(os.Getenv("APP_REGISTER_SECRET"))
}

// SignAppRegister 生成注册签名（客户端与测试用）。
// payload = client + "\n" + email + "\n" + ts
func SignAppRegister(secret, client, email string, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = fmt.Fprintf(mac, "%s\n%s\n%d", client, email, ts)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyAppRegister 校验 App 端免邀请注册凭证。ok=false 时 msg 为对外文案。
func VerifyAppRegister(client, email, sign string, ts int64) (ok bool, msg string) {
	secret := AppRegisterSecret()
	if secret == "" {
		return false, "未开放 App 注册通道"
	}
	client = strings.TrimSpace(client)
	email = strings.TrimSpace(strings.ToLower(email))
	sign = strings.TrimSpace(strings.ToLower(sign))
	if client == "" || email == "" || sign == "" || ts <= 0 {
		return false, "App 注册凭证不完整"
	}
	if _, allowed := allowedAppClients[client]; !allowed {
		return false, "不支持的客户端"
	}
	now := time.Now().Unix()
	skew := int64(appRegisterSkew / time.Second)
	if ts < now-skew || ts > now+skew {
		return false, "注册凭证已过期，请重试"
	}
	expect := SignAppRegister(secret, client, email, ts)
	if !hmac.Equal([]byte(expect), []byte(sign)) {
		return false, "App 注册凭证无效"
	}
	return true, ""
}

// ParseAppRegisterTS 解析客户端时间戳（秒）。
func ParseAppRegisterTS(raw interface{}) int64 {
	switch v := raw.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return n
	default:
		return 0
	}
}

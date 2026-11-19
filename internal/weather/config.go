package weather

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/KyleBing/portal-go/internal/systemconfig"
	"github.com/golang-jwt/jwt/v5"
)

// 和风凭据：优先 JWT（环境变量），否则回退 system_config 里的 API Key。
//
// 环境变量：
//
//	QWEATHER_HOST              API Host（可选；未设则用 system_config.hefeng_weather_api_host）
//	QWEATHER_KID               凭据 ID
//	QWEATHER_ISS               开发者 ID（Q 开头）
//	QWEATHER_SUB               项目 ID
//	QWEATHER_PRIVATE_KEY       PEM 私钥全文（可用 \n 转义）
//	QWEATHER_PRIVATE_KEY_FILE  PEM 私钥文件路径（与上一者二选一）

type creds struct {
	host       string
	kid        string
	iss        string
	sub        string
	privateKey ed25519.PrivateKey
	apiKey     string // fallback
	useJWT     bool
}

var (
	credsMu     sync.Mutex
	cachedCreds *creds
	credsErr    error
	credsLoaded time.Time

	jwtMu     sync.Mutex
	cachedJWT string
	jwtExp    time.Time
)

func loadCreds() (*creds, error) {
	credsMu.Lock()
	defer credsMu.Unlock()
	if cachedCreds != nil && time.Since(credsLoaded) < 30*time.Second {
		return cachedCreds, credsErr
	}

	c := &creds{}
	c.host = sanitizeHost(os.Getenv("QWEATHER_HOST"))
	c.kid = strings.TrimSpace(os.Getenv("QWEATHER_KID"))
	c.iss = strings.TrimSpace(os.Getenv("QWEATHER_ISS"))
	c.sub = strings.TrimSpace(os.Getenv("QWEATHER_SUB"))

	pemStr := strings.TrimSpace(os.Getenv("QWEATHER_PRIVATE_KEY"))
	if pemStr == "" {
		if path := strings.TrimSpace(os.Getenv("QWEATHER_PRIVATE_KEY_FILE")); path != "" {
			b, err := os.ReadFile(path)
			if err != nil {
				credsErr = fmt.Errorf("读取 QWEATHER_PRIVATE_KEY_FILE: %w", err)
				cachedCreds = nil
				credsLoaded = time.Now()
				return nil, credsErr
			}
			pemStr = string(b)
		}
	}
	pemStr = strings.ReplaceAll(pemStr, `\n`, "\n")

	cfg, cfgErr := systemconfig.GetAdminSystemConfig()
	if cfgErr == nil {
		if c.host == "" {
			c.host = sanitizeHost(cfg.HefengWeatherAPIHost)
		}
		c.apiKey = strings.TrimSpace(cfg.HefengWeatherAPIKey)
	}

	if c.kid != "" && c.iss != "" && c.sub != "" && pemStr != "" {
		key, err := parseEd25519PrivateKey(pemStr)
		if err != nil {
			credsErr = fmt.Errorf("解析和风私钥: %w", err)
			cachedCreds = nil
			credsLoaded = time.Now()
			return nil, credsErr
		}
		c.privateKey = key
		c.useJWT = true
	}

	if c.host == "" {
		credsErr = errors.New("和风 API Host 未配置（QWEATHER_HOST 或系统配置 hefeng_weather_api_host）")
		cachedCreds = nil
		credsLoaded = time.Now()
		return nil, credsErr
	}
	if !c.useJWT && c.apiKey == "" {
		credsErr = errors.New("和风未配置：请设置 JWT 环境变量，或在系统配置填写 API Key")
		cachedCreds = nil
		credsLoaded = time.Now()
		return nil, credsErr
	}

	cachedCreds = c
	credsErr = nil
	credsLoaded = time.Now()
	return c, nil
}

func sanitizeHost(h string) string {
	h = strings.TrimSpace(h)
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")
	if i := strings.IndexByte(h, '/'); i >= 0 {
		h = h[:i]
	}
	return strings.TrimSpace(h)
}

func parseEd25519PrivateKey(pemStr string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("无效 PEM")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ed, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("不是 Ed25519 私钥")
	}
	return ed, nil
}

func signQWeatherJWT(c *creds) (string, error) {
	jwtMu.Lock()
	defer jwtMu.Unlock()
	now := time.Now()
	// 缓存到过期前 60s
	if cachedJWT != "" && now.Before(jwtExp.Add(-60*time.Second)) {
		return cachedJWT, nil
	}

	iat := now.Add(-30 * time.Second)
	exp := iat.Add(15 * time.Minute)
	claims := jwt.MapClaims{
		"iss": c.iss,
		"sub": c.sub,
		"iat": iat.Unix(),
		"exp": exp.Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	t.Header["kid"] = c.kid
	delete(t.Header, "typ") // 和风建议不要多余字段；typ 若存在须为 JWT，干脆去掉

	signed, err := t.SignedString(c.privateKey)
	if err != nil {
		return "", err
	}
	cachedJWT = signed
	jwtExp = exp
	return signed, nil
}

// Status 供管理/设备查看是否已配好（不含密钥）。
func Status() map[string]interface{} {
	c, err := loadCreds()
	if err != nil {
		return map[string]interface{}{
			"configured": false,
			"auth":       "none",
			"message":    err.Error(),
		}
	}
	auth := "api_key"
	if c.useJWT {
		auth = "jwt"
	}
	return map[string]interface{}{
		"configured": true,
		"auth":       auth,
		"host":       c.host,
	}
}

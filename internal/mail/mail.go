package mail

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Config 阿里云邮件推送（DirectMail）配置，通过环境变量加载。
type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	AccountName     string
	Region          string
	FromAlias       string
	AppPublicURL    string
	LogoURL         string
}

func (c Config) Enabled() bool {
	return c.AccessKeyID != "" && c.AccessKeySecret != "" && c.AccountName != ""
}

func LoadConfig() Config {
	region := envOr("ALIYUN_DM_REGION", "cn-hangzhou")
	return Config{
		AccessKeyID:     os.Getenv("ALIYUN_DM_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("ALIYUN_DM_ACCESS_KEY_SECRET"),
		AccountName:     os.Getenv("ALIYUN_DM_ACCOUNT_NAME"),
		Region:          region,
		FromAlias:       envOr("ALIYUN_DM_FROM_ALIAS", "日记"),
		AppPublicURL:    strings.TrimRight(envOr("APP_PUBLIC_URL", "http://localhost:3000"), "/"),
		LogoURL:         os.Getenv("MAIL_LOGO_URL"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type Sender struct {
	cfg Config
}

var (
	defaultSender *Sender
	once          sync.Once
)

// Default 返回进程级邮件发送器。
func Default() *Sender {
	once.Do(func() {
		defaultSender = NewSender(LoadConfig())
	})
	return defaultSender
}

func NewSender(cfg Config) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) AppPublicURL() string {
	return s.cfg.AppPublicURL
}

func (s *Sender) SendVerifyEmail(to, verifyURL string) error {
	html, text, err := renderVerifyEmail(s.cfg.LogoURL, verifyURL)
	if err != nil {
		return err
	}
	return s.send(to, "验证你的日记邮箱", html, text)
}

func (s *Sender) SendResetEmail(to, resetURL string) error {
	html, text, err := renderResetEmail(s.cfg.LogoURL, resetURL)
	if err != nil {
		return err
	}
	return s.send(to, "重置日记密码", html, text)
}

func (s *Sender) send(to, subject, html, text string) error {
	if !s.cfg.Enabled() {
		log.Printf("[mail stub] to=%s subject=%s", to, subject)
		return nil
	}

	nonce, err := randomNonce()
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("AccessKeyId", s.cfg.AccessKeyID)
	params.Set("Action", "SingleSendMail")
	params.Set("AccountName", s.cfg.AccountName)
	params.Set("AddressType", "1")
	params.Set("Format", "JSON")
	params.Set("FromAlias", s.cfg.FromAlias)
	params.Set("HtmlBody", html)
	params.Set("TextBody", text)
	params.Set("RegionId", s.cfg.Region)
	params.Set("ReplyToAddress", "false")
	params.Set("SignatureMethod", "HMAC-SHA1")
	params.Set("SignatureNonce", nonce)
	params.Set("SignatureVersion", "1.0")
	params.Set("Subject", subject)
	params.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	params.Set("ToAddress", to)
	params.Set("Version", "2015-11-23")
	params.Set("Signature", rpcSignature("POST", s.cfg.AccessKeySecret, params))

	endpoint := directMailEndpoint(s.cfg.Region)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("mail send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("mail read response: %w", err)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("mail status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		RequestId string `json:"RequestId"`
		Code      string `json:"Code"`
		Message   string `json:"Message"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("mail parse response: %w", err)
	}
	if result.Code != "" {
		return fmt.Errorf("mail api error %s: %s", result.Code, result.Message)
	}

	log.Printf("[mail] sent to %s: %s (requestId=%s)", to, subject, result.RequestId)
	return nil
}

func directMailEndpoint(region string) string {
	if region == "cn-hangzhou" {
		return "https://dm.aliyuncs.com/"
	}
	return fmt.Sprintf("https://dm.%s.aliyuncs.com/", region)
}

func randomNonce() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

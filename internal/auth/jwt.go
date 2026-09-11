package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/KyleBing/portal-go/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenTTL = 30 * 24 * time.Hour
	// renewWindow：剩余有效期短于此窗口时自动续签
	renewWindow = 7 * 24 * time.Hour
	// HeaderRenewedToken 响应头：服务端下发的新 JWT，供客户端写回 localStorage
	HeaderRenewedToken = "X-Access-Token"
)

var (
	ErrMissingSecret = errors.New("JWT_SECRET 未配置")
	ErrInvalidToken  = errors.New("无效的 token")
)

var secret []byte

// Claims is the JWT payload for portal auth.
type Claims struct {
	UID     int64 `json:"uid"`
	GroupID int   `json:"group_id"`
	jwt.RegisteredClaims
}

// Init loads JWT_SECRET from the environment. Call once at process start.
func Init() error {
	s := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if s == "" {
		return ErrMissingSecret
	}
	secret = []byte(s)
	return nil
}

// Issue creates a signed HS256 JWT for the user.
func Issue(u *models.User) (string, error) {
	if len(secret) == 0 {
		return "", ErrMissingSecret
	}
	if u == nil || u.UID <= 0 {
		return "", ErrInvalidToken
	}
	now := time.Now()
	claims := Claims{
		UID:     u.UID,
		GroupID: u.GroupID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret)
}

// Parse verifies a JWT string and returns claims.
func Parse(tokenString string) (*Claims, error) {
	if len(secret) == 0 {
		return nil, ErrMissingSecret
	}
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, ErrInvalidToken
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.UID <= 0 {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ShouldRenew 判断是否接近过期、需要续签。
func ShouldRenew(claims *Claims) bool {
	if claims == nil || claims.ExpiresAt == nil {
		return false
	}
	remaining := time.Until(claims.ExpiresAt.Time)
	return remaining > 0 && remaining < renewWindow
}

// BearerToken extracts the raw JWT from an Authorization header value.
func BearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}

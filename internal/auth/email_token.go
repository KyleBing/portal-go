package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// NewEmailToken 生成原始 token（邮件链接用）与入库的 SHA-256 hex。
func NewEmailToken() (raw string, hash string, err error) {
	var b [32]byte
	if _, err = rand.Read(b[:]); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b[:])
	hash = HashToken(raw)
	return raw, hash, nil
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

package wubi

import (
	"encoding/base64"
	"strings"
	"unicode/utf8"

	"github.com/KyleBing/portal-go/internal/util"
)

// PlainText 把历史存储还原成原文：词库 content 可能是 base64，词条/标题里的 emoji 可能是 \uXXXX 代理对。
func PlainText(s string) string {
	if s == "" {
		return ""
	}
	if decoded, ok := decodeBase64Text(s); ok {
		s = decoded
	}
	return util.UnicodeDecode(s)
}

// decodeBase64Text 只在整段都是 base64、且解码后像 YAML/文本时才还原，避免误伤已经是原文的内容。
func decodeBase64Text(s string) (string, bool) {
	compact := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, strings.TrimSpace(s))
	if len(compact) < 8 || len(compact)%4 != 0 {
		return "", false
	}
	for _, r := range compact {
		if !isBase64Rune(r) {
			return "", false
		}
	}
	raw, err := base64.StdEncoding.DecodeString(compact)
	if err != nil || !utf8.Valid(raw) {
		return "", false
	}
	text := string(raw)
	if !looksLikeText(text) {
		return "", false
	}
	return text, true
}

func isBase64Rune(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '+' || r == '/' || r == '='
}

func looksLikeText(s string) bool {
	if strings.ContainsAny(s, "\n\r:") {
		return true
	}
	for _, r := range s {
		if r > 0x7f {
			return true
		}
	}
	return false
}

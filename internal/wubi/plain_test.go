package wubi

import (
	"encoding/base64"
	"testing"
)

func TestPlainTextBase64YAML(t *testing.T) {
	raw := "你好\ncode: abcd"
	got := PlainText(base64.StdEncoding.EncodeToString([]byte(raw)))
	if got != raw {
		t.Fatalf("got %q", got)
	}
}

func TestPlainTextAlreadyPlain(t *testing.T) {
	raw := "测试🙂\ncode: abcd"
	if got := PlainText(raw); got != raw {
		t.Fatalf("got %q", got)
	}
}

func TestPlainTextSurrogate(t *testing.T) {
	got := PlainText("测试\\uD83D\\uDE42")
	if got != "测试🙂" {
		t.Fatalf("got %q", got)
	}
}

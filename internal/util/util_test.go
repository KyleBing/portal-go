package util

import (
	"strings"
	"testing"
)

func TestUnicodeEncodeDecodeEmoji(t *testing.T) {
	for _, in := range []string{"你好", "测试😀", "🙏后半年"} {
		enc := UnicodeEncode(in)
		dec := UnicodeDecode(enc)
		if dec != in {
			t.Fatalf("roundtrip failed: in=%q enc=%q dec=%q", in, enc, dec)
		}
		// Must store single-backslash form, matching historical Node DB rows.
		if strings.Contains(enc, `\\u`) {
			t.Fatalf("encode produced double-backslash: %q", enc)
		}
		if in != "你好" && !strings.Contains(enc, `\u`) {
			t.Fatalf("encode missing \\u escape: %q", enc)
		}
	}

	// Legacy portal-go rows stored double-backslash escapes.
	if got := UnicodeDecode("测试\\\\uD83D\\\\uDE42"); got != "测试🙂" {
		t.Fatalf("legacy double-backslash decode: got %q", got)
	}

	// Historical Node rows stored single-backslash escapes.
	if got := UnicodeDecode("测试\\uD83D\\uDE42"); got != "测试🙂" {
		t.Fatalf("legacy single-backslash decode: got %q", got)
	}
}

package util

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
)

func DateFormatter(t time.Time, layout string) string {
	if layout == "" || layout == "yyyy-MM-dd hh:mm:ss" {
		return t.Format("2006-01-02 15:04:05")
	}
	r := strings.NewReplacer(
		"yyyy", "2006",
		"MM", "01",
		"dd", "02",
		"hh", "15",
		"mm", "04",
		"ss", "05",
	)
	return t.Format(r.Replace(layout))
}

func NowString() string {
	return DateFormatter(time.Now(), "")
}

// ---------------------------------------------------------------------------
// Faithful ports of JavaScript escape()/unescape() and the diary
// unicodeEncode/unicodeDecode helpers from src/utility.ts. They operate on
// UTF-16 code units to keep byte-for-byte compatibility with data written by
// the original Node service (important for existing diary/qr/map records).
// ---------------------------------------------------------------------------

func isJsEscapeSafe(b uint16) bool {
	switch {
	case b >= 'A' && b <= 'Z':
		return true
	case b >= 'a' && b <= 'z':
		return true
	case b >= '0' && b <= '9':
		return true
	}
	switch b {
	case '@', '*', '_', '+', '-', '.', '/':
		return true
	}
	return false
}

// jsEscape mirrors JavaScript's global escape() function.
func jsEscape(s string) string {
	units := utf16.Encode([]rune(s))
	var b strings.Builder
	for _, u := range units {
		if u < 256 && isJsEscapeSafe(u) {
			b.WriteByte(byte(u))
		} else if u < 256 {
			fmt.Fprintf(&b, "%%%02X", u)
		} else {
			fmt.Fprintf(&b, "%%u%04X", u)
		}
	}
	return b.String()
}

func hexVal(b byte) (int, bool) {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0'), true
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10, true
	case b >= 'A' && b <= 'F':
		return int(b-'A') + 10, true
	}
	return 0, false
}

func parseHex(s string) (uint16, bool) {
	v := 0
	for i := 0; i < len(s); i++ {
		h, ok := hexVal(s[i])
		if !ok {
			return 0, false
		}
		v = v*16 + h
	}
	return uint16(v), true
}

// jsUnescape mirrors JavaScript's global unescape() function. The input is
// expected to be ASCII (it is always produced by jsEscape upstream).
func jsUnescape(s string) string {
	var units []uint16
	i := 0
	for i < len(s) {
		if s[i] == '%' && i+5 < len(s) && (s[i+1] == 'u' || s[i+1] == 'U') {
			if v, ok := parseHex(s[i+2 : i+6]); ok {
				units = append(units, v)
				i += 6
				continue
			}
		}
		if s[i] == '%' && i+2 < len(s) {
			if v, ok := parseHex(s[i+1 : i+3]); ok {
				units = append(units, v)
				i += 3
				continue
			}
		}
		units = append(units, uint16(s[i]))
		i++
	}
	return string(utf16.Decode(units))
}

var encodeSurrogateRe = regexp.MustCompile(`(?i)%u[ed][0-9a-f]{3}`)
var decodeSurrogateRe = regexp.MustCompile(`(?i)%5cu[ed][0-9a-f]{3}`)

// UnicodeEncode mirrors the intended diary unicodeEncode behavior. BMP
// characters (e.g. Chinese) pass through unchanged; astral/surrogate code
// units are stored as \uXXXX so they survive utf8mb3 columns.
//
// Historical note: portal's TS used source.replace('%', '\\\\') which emits
// two backslashes; Node then EscapeMySQLString + SQL string parsing collapsed
// that to a single \uXXXX in the DB. portal-go uses prepared statements, so
// the same double-backslash encode was stored literally and broke round-trips.
// We emit a single backslash to match what is already in the database.
func UnicodeEncode(str string) string {
	if str == "" {
		return ""
	}
	text := jsEscape(str)
	text = encodeSurrogateRe.ReplaceAllStringFunc(text, func(m string) string {
		return strings.Replace(m, "%", "\\", 1)
	})
	return jsUnescape(text)
}

// UnicodeDecode mirrors src/utility.ts unicodeDecode, and also accepts the
// legacy double-backslash \\uXXXX form written by early portal-go builds.
func UnicodeDecode(str string) string {
	if str == "" {
		return ""
	}
	// Collapse \\uXXXX → \uXXXX (portal-go prepared-statement writes).
	// Already-single \uXXXX is unchanged because the needle is two backslashes.
	str = strings.ReplaceAll(str, "\\\\u", "\\u")
	text := jsEscape(str)
	text = decodeSurrogateRe.ReplaceAllStringFunc(text, func(m string) string {
		// replace the leading %5C (case-insensitive) with %
		return "%" + m[3:]
	})
	return jsUnescape(text)
}

// EscapeMySQLString mirrors the escapeMySQLString helper used in diary/starve.
func EscapeMySQLString(str string) string {
	if str == "" {
		return ""
	}
	r := strings.NewReplacer(
		"\\", "\\\\",
		"'", "\\'",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
		"\x00", "\\0",
		"\x1a", "\\Z",
	)
	return r.Replace(str)
}

// UnescapeMySQLString reverses EscapeMySQLString (used when reading diaries).
func UnescapeMySQLString(str string) string {
	if str == "" {
		return ""
	}
	// Order matters: match the original TS replacement order.
	str = strings.ReplaceAll(str, "\\Z", "\x1a")
	str = strings.ReplaceAll(str, "\\0", "\x00")
	str = strings.ReplaceAll(str, "\\t", "\t")
	str = strings.ReplaceAll(str, "\\r", "\r")
	str = strings.ReplaceAll(str, "\\n", "\n")
	str = strings.ReplaceAll(str, "\\\"", "\"")
	str = strings.ReplaceAll(str, "\\'", "'")
	str = strings.ReplaceAll(str, "\\\\", "\\")
	return str
}

func FormatMoney(n float64) float64 {
	return math.Round(n*100) / 100
}

func AtoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func ParseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

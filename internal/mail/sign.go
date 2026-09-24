package mail

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/url"
	"sort"
	"strings"
)

func percentEncode(s string) string {
	replacer := strings.NewReplacer("+", "%20", "*", "%2A", "%7E", "~")
	return replacer.Replace(url.QueryEscape(s))
}

func rpcSignature(method, secret string, params url.Values) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, percentEncode(k)+"="+percentEncode(params.Get(k)))
	}
	canonical := strings.Join(parts, "&")

	stringToSign := method + "&" + percentEncode("/") + "&" + percentEncode(canonical)

	mac := hmac.New(sha1.New, []byte(secret+"&"))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

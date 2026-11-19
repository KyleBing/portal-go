package weather

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 12 * time.Second}

func qwGet(path string, query url.Values) (map[string]interface{}, error) {
	c, err := loadCreds()
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := url.URL{
		Scheme:   "https",
		Host:     c.host,
		Path:     path,
		RawQuery: query.Encode(),
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	// 不要手动设 Accept-Encoding: identity —— 和风仍可能回 gzip，
	// 而 Go 在自定义 Accept-Encoding 时不会自动解压，会得到 \x1f 开头的压缩流。
	// 不设该头时 Transport 会协商 gzip 并透明解压。

	if c.useJWT {
		tok, err := signQWeatherJWT(c)
		if err != nil {
			return nil, fmt.Errorf("签发和风 JWT: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tok)
	} else {
		req.Header.Set("X-QW-Api-Key", c.apiKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("和风 HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	body, err = maybeGunzip(body)
	if err != nil {
		return nil, fmt.Errorf("和风 gzip 解压失败: %w", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("和风 JSON 解析失败: %w", err)
	}
	if code, _ := out["code"].(string); code != "" && code != "200" {
		return out, fmt.Errorf("和风业务码 %s", code)
	}
	return out, nil
}

// maybeGunzip：万一响应体仍是 gzip 魔数，手动解开（兜底）。
func maybeGunzip(body []byte) ([]byte, error) {
	if len(body) >= 2 && body[0] == 0x1f && body[1] == 0x8b {
		zr, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		defer zr.Close()
		return io.ReadAll(io.LimitReader(zr, 2<<20))
	}
	return body, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func FetchNow(location string) (map[string]interface{}, error) {
	q := url.Values{}
	q.Set("location", location)
	q.Set("lang", "zh")
	return qwGet("/v7/weather/now", q)
}

func FetchDaily(location string, days string) (map[string]interface{}, error) {
	switch days {
	case "3d", "7d", "10d", "15d", "30d":
	default:
		return nil, fmt.Errorf("不支持的 days=%s（可用 3d/7d/10d/15d/30d）", days)
	}
	q := url.Values{}
	q.Set("location", location)
	q.Set("lang", "zh")
	return qwGet("/v7/weather/"+days, q)
}

func FetchGeo(location string, number string, rangeCN bool) (map[string]interface{}, error) {
	q := url.Values{}
	q.Set("location", location)
	q.Set("lang", "zh")
	if number == "" {
		number = "10"
	}
	q.Set("number", number)
	if rangeCN {
		q.Set("range", "cn")
	}
	return qwGet("/geo/v2/city/lookup", q)
}

package weather

import (
	"net/url"
	"path"
	"strings"
	"unicode"

	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/gin-gonic/gin"
)

// 和风官方路径前缀（GET 原样转发 query）。含仍在用的 v7（LocationID）与现行 v1（经纬度）。
// 文档：https://dev.qweather.com/docs/api/
var officialPrefixes = []string{
	"/v7/",
	"/geo/",
	"/weather/v1/",
	"/airquality/",
	"/weatheralert/",
	"/solarradiation/",
}

const maxOfficialPath = 160

func officialPathFromRequest(raw string) string {
	p := strings.TrimSpace(raw)
	if i := strings.IndexByte(p, '?'); i >= 0 {
		p = p[:i]
	}
	p = strings.TrimPrefix(p, "/portal")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if strings.Contains(p, "..") {
		return ""
	}
	return path.Clean(p)
}

func isOfficialQWeatherPath(p string) bool {
	if p == "" || len(p) > maxOfficialPath {
		return false
	}
	if strings.Contains(p, "..") || strings.Contains(p, "//") || strings.Contains(p, "@") || strings.Contains(p, "\\") {
		return false
	}
	okPrefix := false
	for _, pre := range officialPrefixes {
		if strings.HasPrefix(p, pre) {
			okPrefix = true
			break
		}
	}
	if !okPrefix {
		return false
	}
	for _, r := range p {
		if r <= unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r) || r == '/' || r == '.' || r == '-' || r == '_') {
			continue
		}
		return false
	}
	return true
}

// handleOfficialProxy 透传和风官方路径：GET /portal/v7/weather/24h?location=… → 和风同 path+query。
func handleOfficialProxy(c *gin.Context) {
	if _, ok := middleware.RequireUser(c); !ok {
		return
	}
	p := officialPathFromRequest(c.Request.URL.Path)
	if !isOfficialQWeatherPath(p) {
		response.Error(c, nil, "不允许代理的路径（须为和风官方前缀 /v7 /geo /weather/v1 /airquality /weatheralert /solarradiation）")
		return
	}
	q := c.Request.URL.Query()
	if q == nil {
		q = url.Values{}
	}
	data, err := qwGet(p, q)
	if err != nil {
		response.Error(c, data, err.Error())
		return
	}
	response.Success(c, data, "ok")
}

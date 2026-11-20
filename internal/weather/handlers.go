package weather

import (
	"strings"

	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/gin-gonic/gin"
)

// Register 挂到 portal 根组：官方和风路径原样代理，另保留 /weather/* 短别名。
func Register(rg *gin.RouterGroup) {
	rg.GET("/v7/*path", handleOfficialProxy)
	rg.GET("/geo/*path", handleOfficialProxy)
	rg.GET("/airquality/*path", handleOfficialProxy)
	rg.GET("/weatheralert/*path", handleOfficialProxy)
	rg.GET("/solarradiation/*path", handleOfficialProxy)

	r := rg.Group("/weather")
	r.GET("/status", handleStatus)
	r.GET("/now", handleNow)
	r.GET("/hourly", handleHourly)
	r.GET("/daily", handleDaily)
	r.GET("/geo", handleGeo)
	r.GET("/v1/*path", handleOfficialProxy)
}

func handleStatus(c *gin.Context) {
	if _, ok := middleware.RequireUser(c); !ok {
		return
	}
	response.Success(c, Status(), "ok")
}

func handleNow(c *gin.Context) {
	if _, ok := middleware.RequireUser(c); !ok {
		return
	}
	loc := strings.TrimSpace(c.Query("location"))
	if loc == "" {
		response.Error(c, nil, "缺少 location（LocationID 或 lon,lat）")
		return
	}
	data, err := FetchNow(loc)
	if err != nil {
		response.Error(c, data, err.Error())
		return
	}
	response.Success(c, data, "ok")
}

func handleHourly(c *gin.Context) {
	if _, ok := middleware.RequireUser(c); !ok {
		return
	}
	loc := strings.TrimSpace(c.Query("location"))
	if loc == "" {
		response.Error(c, nil, "缺少 location")
		return
	}
	hours := strings.TrimSpace(c.Query("hours"))
	if hours == "" {
		hours = "24h"
	}
	// 允许传 24 / 72 / 168
	if !strings.HasSuffix(hours, "h") {
		hours = hours + "h"
	}
	data, err := FetchHourly(loc, hours)
	if err != nil {
		response.Error(c, data, err.Error())
		return
	}
	response.Success(c, data, "ok")
}

func handleDaily(c *gin.Context) {
	if _, ok := middleware.RequireUser(c); !ok {
		return
	}
	loc := strings.TrimSpace(c.Query("location"))
	if loc == "" {
		response.Error(c, nil, "缺少 location")
		return
	}
	days := strings.TrimSpace(c.Query("days"))
	if days == "" {
		days = "7d"
	}
	// 允许传 7 / 15
	if !strings.HasSuffix(days, "d") {
		days = days + "d"
	}
	data, err := FetchDaily(loc, days)
	if err != nil {
		response.Error(c, data, err.Error())
		return
	}
	response.Success(c, data, "ok")
}

func handleGeo(c *gin.Context) {
	if _, ok := middleware.RequireUser(c); !ok {
		return
	}
	loc := strings.TrimSpace(c.Query("location"))
	if loc == "" {
		response.Error(c, nil, "缺少 location（城市名）")
		return
	}
	number := strings.TrimSpace(c.Query("number"))
	rangeCN := c.Query("range") != "all"
	data, err := FetchGeo(loc, number, rangeCN)
	if err != nil {
		response.Error(c, data, err.Error())
		return
	}
	response.Success(c, data, "ok")
}

package server

import (
	"net/http"
	"path/filepath"

	"github.com/KyleBing/portal-go/internal/bankcard"
	"github.com/KyleBing/portal-go/internal/bill"
	"github.com/KyleBing/portal-go/internal/category"
	"github.com/KyleBing/portal-go/internal/diary"
	"github.com/KyleBing/portal-go/internal/file"
	"github.com/KyleBing/portal-go/internal/imageqiniu"
	initdb "github.com/KyleBing/portal-go/internal/init"
	"github.com/KyleBing/portal-go/internal/invitation"
	"github.com/KyleBing/portal-go/internal/maproute"
	"github.com/KyleBing/portal-go/internal/mappointer"
	"github.com/KyleBing/portal-go/internal/qr"
	"github.com/KyleBing/portal-go/internal/setup"
	"github.com/KyleBing/portal-go/internal/starve"
	"github.com/KyleBing/portal-go/internal/starvenew"
	"github.com/KyleBing/portal-go/internal/statistic"
	"github.com/KyleBing/portal-go/internal/systemconfig"
	"github.com/KyleBing/portal-go/internal/thumbsup"
	"github.com/KyleBing/portal-go/internal/user"
	"github.com/KyleBing/portal-go/internal/userconfig"
	"github.com/KyleBing/portal-go/internal/wubi"
	"github.com/gin-gonic/gin"
)

// New 构建 gin 引擎并注册所有模块的路由
func New() *gin.Engine {
	// setup 的初始化回调（避免包循环）
	setup.InitFn = initdb.HandleInitJSON

	r := gin.Default()
	r.MaxMultipartMemory = 50 << 20 // 50 MiB

	// 健康检查 / 首页
	r.GET("/", healthHandler)

	// 静态资源：前端管理后台
	r.Static("/manager", filepath.Join(setup.ProjectRoot(), "web", "manager", "dist"))

	// 在根路径与 /portal 前缀下分别注册所有模块
	registerAll(&r.RouterGroup)
	portal := r.Group("/portal")
	portal.GET("/", healthHandler)
	registerAll(portal)

	return r
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Portal API is running",
		"title":   "Portal for Diary",
	})
}

// registerAll 在给定的路由组下注册所有业务模块
func registerAll(rg *gin.RouterGroup) {
	user.Register(rg.Group("/user"))
	initdb.Register(rg.Group("/init"))
	setup.Register(rg.Group("/setup"))
	systemconfig.Register(rg.Group("/system-config"))
	userconfig.Register(rg.Group("/user-config"))
	invitation.Register(rg.Group("/invitation"))

	qr.RegisterFront(rg.Group("/qr-front"))
	qr.RegisterManager(rg.Group("/qr-manager"))

	maproute.Register(rg.Group("/map-route"))
	mappointer.Register(rg.Group("/map-pointer"))

	statistic.Register(rg.Group("/statistic"))

	diary.Register(rg.Group("/diary"))
	category.Register(rg.Group("/diary-category"))
	bankcard.Register(rg.Group("/bank-card"))
	bill.Register(rg.Group("/bill"))

	thumbsup.Register(rg.Group("/thumbs-up"))
	file.Register(rg.Group("/file-manager"))
	imageqiniu.Register(rg.Group("/image-qiniu"))

	wubi.RegisterDict(rg.Group("/wubi/dict"))
	wubi.RegisterWord(rg.Group("/wubi/word"))
	wubi.RegisterCategory(rg.Group("/wubi/category"))

	starve.Register(rg.Group("/starve"))
	starvenew.Register(rg.Group("/starve-new"))
}

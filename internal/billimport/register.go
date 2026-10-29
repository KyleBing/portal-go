package billimport

import "github.com/gin-gonic/gin"

// Register 挂载 /bill-import 路由
func Register(r *gin.RouterGroup) {
	r.GET("/aliases", handleGetAliases)
	r.PUT("/aliases", handlePutAliases)
	r.POST("/aliases/seed", handleSeedAliases)
	r.POST("/upload", handleUpload)
	r.GET("/preview", handlePreview)
	r.POST("/apply", handleApply)
	r.GET("/batches", handleListBatches)
	r.GET("/orders", handleListOrders)
}

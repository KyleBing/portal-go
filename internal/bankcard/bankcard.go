package bankcard

import (
	"github.com/KyleBing/portal-go/internal/apihelper"
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

// Register mounts /bank-card routes.
func Register(r *gin.RouterGroup) {
	g := r
	g.GET("", handleGet)
}

func handleGet(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", errMsg)
		return
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	data, err := apihelper.QueryMap(diary, `select * from diaries where title = '我的银行卡列表' and uid = ?`, user.UID)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	if data != nil {
		util.UpdateUserLastLoginTime(user.UID)
		response.Success(c, util.UnicodeDecode(apihelper.MapStr(data, "content")), "")
	} else {
		response.Success(c, "", "未保存任何银行卡信息")
	}
}

package user

import (
	"database/sql"
	"strings"

	"github.com/KyleBing/portal-go/internal/apihelper"
	"github.com/KyleBing/portal-go/internal/auth"
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/models"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/systemconfig"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const dataName = "用户"
const table = "users"

func diaryDB() (*sql.DB, error) { return db.Open(db.Diary) }

// Register mounts /user routes.
func Register(r *gin.RouterGroup) {
	g := r
	g.POST("/register", handleRegister)
	g.POST("/list", handleList)
	g.GET("/detail", handleDetail)
	g.GET("/avatar", handleAvatar)
	g.POST("/add", handleAdd)
	g.PUT("/set-profile", handleSetProfile)
	g.PUT("/modify", handleModify)
	g.DELETE("/delete", handleDelete)
	g.POST("/login", handleLogin)
	g.PUT("/change-password", handleChangePassword)
	g.DELETE("/destroy-account", handleDestroyAccount)
}

func handleRegister(c *gin.Context) {
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "数据库请求出错")
		return
	}
	body := apihelper.Body(c)

	var userCount int64
	if err := diary.QueryRow(`select count(*) as userCount from ` + table).Scan(&userCount); err != nil {
		response.Error(c, err.Error(), "数据库请求出错")
		return
	}
	if userCount == 0 {
		registerUser(c, diary, body)
		return
	}

	cfg, _ := systemconfig.GetAdminSystemConfig()
	invitationCode := strings.TrimSpace(apihelper.S(body, "invitationCode"))
	if cfg.InvitationCode != "" && invitationCode == cfg.InvitationCode {
		registerUser(c, diary, body)
		return
	}

	inv, err := apihelper.QueryMap(diary, `select * from invitations where id = ?`, invitationCode)
	if err != nil {
		response.Error(c, err.Error(), "数据库请求出错")
		return
	}
	if inv != nil {
		if inv["binding_uid"] != nil {
			response.Error(c, "", "邀请码已被使用")
		} else {
			registerUser(c, diary, body)
		}
	} else {
		response.Error(c, "", "邀请码无效")
	}
}

func checkEmailOrUsernameExist(diary *sql.DB, email, username string) ([]map[string]interface{}, error) {
	return apihelper.QueryMaps(diary, `select * from `+table+` where email=? or username =?`, email, username)
}

func registerUser(c *gin.Context, diary *sql.DB, body map[string]interface{}) {
	email := apihelper.S(body, "email")
	username := apihelper.S(body, "username")
	exist, err := checkEmailOrUsernameExist(diary, email, username)
	if err != nil {
		response.Error(c, err.Error(), "查询出错")
		return
	}
	if len(exist) > 0 {
		response.Error(c, "", "邮箱或用户名已被注册")
		return
	}

	var userCount int64
	if err := diary.QueryRow(`select count(*) as userCount from ` + table).Scan(&userCount); err != nil {
		response.Error(c, "", "检查用户数量失败")
		return
	}
	isFirstUser := userCount == 0
	groupID := 2
	if isFirstUser {
		groupID = 1
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(apihelper.S(body, "password")), 10)
	if err != nil {
		response.Error(c, "", "注册失败")
		return
	}
	now := util.NowString()
	_, err = diary.Exec(`insert into `+table+`(email, nickname, username, password, register_time, last_visit_time, comment, wx, phone, homepage, gaode, group_id)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		email, apihelper.S(body, "nickname"), apihelper.S(body, "username"), string(hash), now, now,
		apihelper.S(body, "comment"), apihelper.S(body, "wx"), apihelper.S(body, "phone"),
		apihelper.S(body, "homepage"), apihelper.S(body, "gaode"), groupID)
	if err != nil {
		response.Error(c, "", "注册失败")
		return
	}

	invitationCode := apihelper.S(body, "invitationCode")
	if invitationCode != "" {
		var uid int64
		_ = diary.QueryRow(`select uid from `+table+` where email=?`, email).Scan(&uid)
		_, ierr := diary.Exec(`update invitations set binding_uid = ?, date_register = ? where id = ?`, uid, now, invitationCode)
		if ierr != nil {
			msg := "注册成功，邀请码信息更新失败"
			if isFirstUser {
				msg = "注册成功！您已成为系统管理员。邀请码信息更新失败"
			}
			response.Error(c, "", msg)
			return
		}
	}
	msg := "注册成功"
	if isFirstUser {
		msg = "注册成功！您已成为系统管理员。"
	}
	response.Success(c, "", msg)
}

func handleList(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", errMsg)
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	body := apihelper.Body(c)
	pageNo := int(apihelper.I(body, "pageNo"))
	pageSize := int(apihelper.I(body, "pageSize"))
	start := apihelper.PageStart(pageNo, pageSize)

	var list []map[string]interface{}
	var total int64
	if user.IsAdmin() {
		list, err = apihelper.QueryMaps(diary, `SELECT * from `+table+` limit ?, ?`, start, pageSize)
		if err == nil {
			err = diary.QueryRow(`select count(*) as sum from ` + table).Scan(&total)
		}
	} else {
		list, err = apihelper.QueryMaps(diary, `SELECT * from `+table+` where uid = ? limit ?, ?`, user.UID, start, pageSize)
		if err == nil {
			err = diary.QueryRow(`select count(*) as sum from `+table+` where uid = ?`, user.UID).Scan(&total)
		}
	}
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	// 列表不返回密码哈希
	for _, row := range list {
		delete(row, "password")
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{
		"list": list,
		"pager": gin.H{"pageSize": pageSize, "pageNo": pageNo, "total": total},
	}, "请求成功")
}

// handleDetail mirrors the (quirky) original: queries qrs by hash.
func handleDetail(c *gin.Context) {
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	hash := c.Query("hash")
	data, err := apihelper.QueryMap(diary, `select * from qrs where hash = ?`, hash)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	if data == nil {
		response.Error(c, "", "查无此码")
		return
	}
	data["message"] = util.UnicodeDecode(apihelper.MapStr(data, "message"))
	data["description"] = util.UnicodeDecode(apihelper.MapStr(data, "description"))
	if apihelper.MapInt(data, "is_public") == 1 {
		response.Success(c, data, "")
		return
	}
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", errMsg)
		return
	}
	if user.UID == apihelper.MapInt(data, "uid") {
		util.UpdateUserLastLoginTime(user.UID)
		response.Success(c, data, "")
	} else {
		response.Error(c, "", "当前用户无权查看该 QR ：请求用户 ID 与 QR 归属不匹配")
	}
}

func handleAvatar(c *gin.Context) {
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	data, err := apihelper.QueryMap(diary, `select avatar from `+table+` where email = ?`, c.Query("email"))
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	response.Success(c, data, "")
}

func handleAdd(c *gin.Context) {
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	body := apihelper.Body(c)
	exist, err := checkEmailOrUsernameExist(diary, apihelper.S(body, "email"), apihelper.S(body, "username"))
	if err != nil {
		response.Error(c, err.Error(), "查询出错")
		return
	}
	if len(exist) > 0 {
		response.Error(c, "", "邮箱或用户名已被注册")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(apihelper.S(body, "password")), 10)
	if err != nil {
		response.Error(c, "", "用户添加失败")
		return
	}
	now := util.NowString()
	_, err = diary.Exec(`insert into `+table+`(email, nickname, username, password, register_time, last_visit_time, comment, wx, phone, homepage, gaode, group_id)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		apihelper.S(body, "email"), apihelper.S(body, "nickname"), apihelper.S(body, "username"), string(hash), now, now,
		apihelper.S(body, "comment"), apihelper.S(body, "wx"), apihelper.S(body, "phone"),
		apihelper.S(body, "homepage"), apihelper.S(body, "gaode"), apihelper.S(body, "group_id"))
	if err != nil {
		response.Error(c, err.Error(), "用户添加失败")
		return
	}
	response.Success(c, "", "用户添加成功")
}

func handleSetProfile(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	if systemconfig.IsConfiguredDemoAccountEmail(user.Email) {
		response.Error(c, "", "演示帐户不允许修改资料哦")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "修改失败")
		return
	}
	body := apihelper.Body(c)
	_, err = diary.Exec(`update `+table+` set nickname=?, phone=?, avatar=?, city=?, geolocation=? WHERE uid = ?`,
		apihelper.S(body, "nickname"), apihelper.S(body, "phone"), apihelper.S(body, "avatar"),
		apihelper.S(body, "city"), apihelper.S(body, "geolocation"), user.UID)
	if err != nil {
		response.Error(c, err.Error(), "修改失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	newUser, _ := middleware.VerifyAuthorization(c)
	response.Success(c, newUser, "修改成功")
}

func handleModify(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	body := apihelper.Body(c)
	targetUID := apihelper.I(body, "uid")
	if !user.IsAdmin() && user.UID != targetUID {
		response.Error(c, "", "你无权操作该用户信息")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), "修改失败")
		return
	}
	_, err = diary.Exec(`update `+table+` set email=?, nickname=?, username=?, comment=?, wx=?, phone=?, homepage=?, gaode=?, group_id=? WHERE uid=?`,
		apihelper.S(body, "email"), apihelper.S(body, "nickname"), apihelper.S(body, "username"),
		apihelper.S(body, "comment"), apihelper.S(body, "wx"), apihelper.S(body, "phone"),
		apihelper.S(body, "homepage"), apihelper.S(body, "gaode"), apihelper.S(body, "group_id"), apihelper.S(body, "uid"))
	if err != nil {
		response.Error(c, err.Error(), "修改失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "修改成功")
}

func handleDelete(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), dataName+"删除失败")
		return
	}
	body := apihelper.Body(c)
	apihelper.OperateNoReturn(c, diary, user.UID, dataName, "删除", `DELETE from `+table+` WHERE uid=?`, apihelper.S(body, "uid"))
}

func handleLogin(c *gin.Context) {
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, "", err.Error())
		return
	}
	body := apihelper.Body(c)
	data, err := apihelper.QueryMap(diary, `select * from `+table+` where email = ?`, apihelper.S(body, "email"))
	if err != nil {
		response.Error(c, "", err.Error())
		return
	}
	if data == nil {
		response.Error(c, "", "无此用户")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(apihelper.MapStr(data, "password")), []byte(apihelper.S(body, "password"))) != nil {
		response.Error(c, "", "用户名或密码错误")
		return
	}
	uid := apihelper.MapInt(data, "uid")
	u := &models.User{
		UID:     uid,
		GroupID: int(apihelper.MapInt(data, "group_id")),
	}
	token, err := auth.Issue(u)
	if err != nil {
		response.Error(c, "", "签发 token 失败")
		return
	}
	delete(data, "password")
	data["token"] = token
	util.UpdateUserLastLoginTime(uid)
	response.Success(c, data, "登录成功")
}

func handleChangePassword(c *gin.Context) {
	body := apihelper.Body(c)
	if apihelper.S(body, "password") == "" {
		response.Error(c, "", "参数错误：password 未定义")
		return
	}
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	if systemconfig.IsConfiguredDemoAccountEmail(user.Email) {
		response.Error(c, "", "演示帐户密码不允许修改")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, err.Error(), dataName+"修改密码失败")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(apihelper.S(body, "password")), 10)
	if err != nil {
		response.Error(c, err.Error(), dataName+"修改密码失败")
		return
	}
	apihelper.OperateReturnID(c, diary, user.UID, dataName, "修改密码", `update `+table+` set password = ? where email=?`, string(hash), user.Email)
}

func handleDestroyAccount(c *gin.Context) {
	user, errMsg := middleware.VerifyAuthorization(c)
	if errMsg != "" {
		response.Error(c, "null", errMsg)
		return
	}
	if systemconfig.IsConfiguredDemoAccountEmail(user.Email) {
		response.Error(c, "", "演示帐户不允许执行此操作")
		return
	}
	diary, err := diaryDB()
	if err != nil {
		response.Error(c, "", "beginTransaction: 事务执行失败，已回滚")
		return
	}
	tx, err := diary.Begin()
	if err != nil {
		response.Error(c, "", "beginTransaction: 事务执行失败，已回滚")
		return
	}
	stmts := []string{
		`delete from diaries where uid = ?`,
		`delete from invitations where binding_uid = ?`,
		`delete from map_pointer where uid = ?`,
		`delete from map_route where uid = ?`,
		`delete from qrs where uid = ?`,
		`delete from ` + table + ` where uid = ?`,
	}
	for _, s := range stmts {
		if _, err := tx.Exec(s, user.UID); err != nil {
			_ = tx.Rollback()
			response.Error(c, err.Error(), "query: 事务执行失败，已回滚")
			return
		}
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		response.Error(c, err.Error(), "transaction.commit: 事务执行失败，已回滚")
		return
	}
	response.Success(c, nil, "事务执行成功")
}

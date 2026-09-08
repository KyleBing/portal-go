package qr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/middleware"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/KyleBing/portal-go/internal/util"
	"github.com/gin-gonic/gin"
)

const dbName = db.Diary
const currentTable = "qrs"

// RegisterFront 注册二维码前台接口（/qr-front）
func RegisterFront(r *gin.RouterGroup) {
	r.GET("/", handleFront)
}

// RegisterManager 注册二维码后台接口（/qr-manager）
func RegisterManager(r *gin.RouterGroup) {
	r.GET("/list", handleList)
	r.GET("/detail", handleDetail)
	r.POST("/add", handleAdd)
	r.PUT("/modify", handleModify)
	r.DELETE("/delete", handleDelete)
	r.POST("/clear-visit-count", handleClearVisitCount)
}

func handleFront(c *gin.Context) {
	hash := c.Query("hash")
	diary, _ := db.Open(dbName)
	row, err := db.QueryMap(diary, `
		select qrs.hash, qrs.is_public, qrs.is_show_phone, qrs.message, qrs.car_name, qrs.car_plate,
		       qrs.car_desc, qrs.is_show_car, qrs.is_show_wx, qrs.wx_code_img, qrs.description,
		       qrs.is_show_homepage, qrs.is_show_gaode, qrs.date_init, qrs.visit_count, qrs.imgs, qrs.car_type,
		       users.phone, users.wx, users.homepage, users.uid, users.nickname, users.username
		from qrs left join users on qrs.uid = users.uid
		where qrs.hash = ? and is_public = 1`, hash)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	if row == nil {
		response.Error(c, "", "查无此码")
		return
	}
	row["message"] = util.UnicodeDecode(asString(row["message"]))
	row["description"] = util.UnicodeDecode(asString(row["description"]))

	hashList, err := db.QueryMaps(diary, `select hash, car_name, car_plate, imgs from qrs where uid = ? and is_public = 1 and hash != ?`, asInt(row["uid"]), hash)
	if err != nil {
		response.Error(c, "", "获取同用户其它码失败")
		return
	}
	response.Success(c, gin.H{"dataQr": row, "hashList": hashList}, "获取成功")
	// 访问次数 +1
	_, _ = diary.Exec(`update `+currentTable+` set visit_count = visit_count + 1 where hash = ?`, hash)
}

func handleList(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权查看 QR 列表：用户信息错误")
		return
	}
	base := `select qrs.hash, qrs.is_public, qrs.is_show_phone, qrs.message, qrs.car_name, qrs.car_plate,
		qrs.car_desc, qrs.is_show_car, qrs.is_show_wx, qrs.wx_code_img, qrs.description,
		qrs.is_show_homepage, qrs.is_show_gaode, qrs.date_init, qrs.visit_count, qrs.imgs, qrs.car_type,
		users.phone, users.wx, users.uid, users.nickname, users.username
		from qrs left join users on qrs.uid = users.uid`

	var filters []string
	var args []interface{}
	if !user.IsAdmin() {
		filters = append(filters, "qrs.uid = ?")
		args = append(args, user.UID)
	}
	for _, kw := range parseJSONStringArray(c.Query("keywords")) {
		filters = append(filters, "( message like ? ESCAPE '/'  or description like ? ESCAPE '/')")
		like := "%" + util.UnicodeEncode(kw) + "%"
		args = append(args, like, like)
	}
	filterSQL := ""
	if len(filters) > 0 {
		filterSQL = " where " + strings.Join(filters, " and ")
	}

	pageNo := util.AtoiDefault(c.Query("pageNo"), 1)
	pageSize := util.AtoiDefault(c.Query("pageSize"), 20)
	startPoint := (pageNo - 1) * pageSize
	args = append(args, startPoint, pageSize)

	diary, _ := db.Open(dbName)
	rows, err := db.QueryMaps(diary, base+filterSQL+" order by date_init desc limit ?, ?", args...)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	for _, row := range rows {
		row["message"] = util.UnicodeDecode(asString(row["message"]))
		row["description"] = util.UnicodeDecode(asString(row["description"]))
	}
	response.Success(c, rows, "请求成功")
}

func handleDetail(c *gin.Context) {
	hash := c.Query("hash")
	diary, _ := db.Open(dbName)
	row, err := db.QueryMap(diary, `select * from `+currentTable+` where hash = ?`, hash)
	if err != nil {
		response.Error(c, err.Error(), err.Error())
		return
	}
	if row == nil {
		response.Error(c, "", "查无此码")
		return
	}
	row["message"] = util.UnicodeDecode(asString(row["message"]))
	row["description"] = util.UnicodeDecode(asString(row["description"]))
	if asInt(row["is_public"]) == 1 {
		response.Success(c, row, "")
		return
	}
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "当前用户无权查看该 QR ：用户信息错误")
		return
	}
	if user.UID == asInt(row["uid"]) {
		util.UpdateUserLastLoginTime(user.UID)
		response.Success(c, row, "")
	} else {
		response.Error(c, "", "当前用户无权查看该 QR ：请求用户 ID 与 QR 归属不匹配")
	}
}

type qrBody struct {
	Hash           string `json:"hash"`
	IsPublic       int    `json:"is_public"`
	IsShowPhone    int    `json:"is_show_phone"`
	Message        string `json:"message"`
	Description    string `json:"description"`
	CarName        string `json:"car_name"`
	CarPlate       string `json:"car_plate"`
	CarDesc        string `json:"car_desc"`
	IsShowCar      int    `json:"is_show_car"`
	WxCodeImg      string `json:"wx_code_img"`
	IsShowWx       int    `json:"is_show_wx"`
	IsShowHomepage int    `json:"is_show_homepage"`
	IsShowGaode    int    `json:"is_show_gaode"`
	VisitCount     int    `json:"visit_count"`
	UID            int64  `json:"uid"`
	Imgs           string `json:"imgs"`
	CarType        int    `json:"car_type"`
}

func handleAdd(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	var body qrBody
	_ = c.ShouldBindJSON(&body)
	hash := strings.ToLower(body.Hash)
	diary, _ := db.Open(dbName)
	exist, _ := db.QueryMaps(diary, `select * from `+currentTable+` where hash=?`, hash)
	if len(exist) > 0 {
		response.Error(c, "", fmt.Sprintf("已存在名为 %s 的记录", body.Hash))
		return
	}
	now := util.NowString()
	_, err := diary.Exec(`insert into `+currentTable+`(hash, is_public, is_show_phone, message, description, car_name, car_plate, car_desc, is_show_car, wx_code_img, is_show_wx, is_show_homepage, is_show_gaode, date_modify, date_init, visit_count, uid, imgs, car_type)
		values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		hash, body.IsPublic, body.IsShowPhone, util.UnicodeEncode(body.Message), util.UnicodeEncode(body.Description),
		body.CarName, body.CarPlate, body.CarDesc, body.IsShowCar, body.WxCodeImg, body.IsShowWx,
		body.IsShowHomepage, body.IsShowGaode, now, now, body.VisitCount, user.UID, body.Imgs, body.CarType)
	if err != nil {
		response.Error(c, err.Error(), "二维码添加失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": 0}, "添加成功")
}

func handleModify(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	var body qrBody
	_ = c.ShouldBindJSON(&body)
	now := util.NowString()
	diary, _ := db.Open(dbName)
	_, err := diary.Exec(`update `+currentTable+` set is_public=?, is_show_phone=?, message=?, description=?, car_name=?, car_plate=?, car_desc=?, is_show_car=?, is_show_wx=?, wx_code_img=?, is_show_homepage=?, is_show_gaode=?, date_modify=?, visit_count=?, uid=?, imgs=?, car_type=? WHERE hash=?`,
		body.IsPublic, body.IsShowPhone, util.UnicodeEncode(body.Message), util.UnicodeEncode(body.Description),
		body.CarName, body.CarPlate, body.CarDesc, body.IsShowCar, body.IsShowWx, body.WxCodeImg,
		body.IsShowHomepage, body.IsShowGaode, now, body.VisitCount, body.UID, body.Imgs, body.CarType, body.Hash)
	if err != nil {
		response.Error(c, err.Error(), "二维码修改失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "修改成功")
}

func handleDelete(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		Hash string `json:"hash"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	var err error
	if user.IsAdmin() {
		_, err = diary.Exec(`DELETE from `+currentTable+` WHERE hash=?`, body.Hash)
	} else {
		_, err = diary.Exec(`DELETE from `+currentTable+` WHERE hash=? and uid=?`, body.Hash, user.UID)
	}
	if err != nil {
		response.Error(c, err.Error(), "二维码删除失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, nil, "删除成功")
}

func handleClearVisitCount(c *gin.Context) {
	user, msg := middleware.VerifyAuthorization(c)
	if msg != "" {
		response.Error(c, "", "无权操作")
		return
	}
	if !user.IsAdmin() {
		response.Error(c, "", "无权操作")
		return
	}
	var body struct {
		Hash string `json:"hash"`
	}
	_ = c.ShouldBindJSON(&body)
	diary, _ := db.Open(dbName)
	_, err := diary.Exec(`update `+currentTable+` set visit_count = 0 where hash = ?`, body.Hash)
	if err != nil {
		response.Error(c, err.Error(), "计数清零失败")
		return
	}
	util.UpdateUserLastLoginTime(user.UID)
	response.Success(c, gin.H{"id": 0}, "计数清零成功")
}

func parseJSONStringArray(s string) []string {
	if s == "" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil
	}
	return arr
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprintf("%v", s)
	}
}

func asInt(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case []byte:
		var x int64
		fmt.Sscan(string(n), &x)
		return x
	case string:
		var x int64
		fmt.Sscan(n, &x)
		return x
	}
	return 0
}

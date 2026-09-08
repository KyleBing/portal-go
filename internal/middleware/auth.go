package middleware

import (
	"database/sql"
	"strconv"

	"github.com/KyleBing/portal-go/internal/db"
	"github.com/KyleBing/portal-go/internal/models"
	"github.com/KyleBing/portal-go/internal/response"
	"github.com/gin-gonic/gin"
)

const ContextUserKey = "authUser"

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, errMsg := VerifyAuthorization(c)
		if errMsg != "" {
			response.Error(c, "", errMsg)
			c.Abort()
			return
		}
		c.Set(ContextUserKey, user)
		c.Next()
	}
}

func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := VerifyAuthorization(c)
		if user != nil {
			c.Set(ContextUserKey, user)
		}
		c.Next()
	}
}

func GetUser(c *gin.Context) *models.User {
	v, ok := c.Get(ContextUserKey)
	if !ok {
		return nil
	}
	u, _ := v.(*models.User)
	return u
}

func VerifyAuthorization(c *gin.Context) (*models.User, string) {
	token := c.GetHeader("Diary-Token")
	if token == "" {
		token = c.Query("token")
	}
	uidStr := c.GetHeader("Diary-Uid")
	if token == "" {
		return nil, "无 token"
	}
	if uidStr == "" {
		return nil, "程序已升级，请关闭所有相关窗口，再重新访问该网站"
	}
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		return nil, "身份验证失败：查无此人"
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		return nil, "mysql: 获取身份信息错误"
	}
	user, err := ScanUser(diary.QueryRow(`SELECT uid,email,nickname,username,password,register_time,last_visit_time,comment,wx,phone,homepage,gaode,group_id,count_diary,count_dict,count_qr,count_words,count_map_route,count_map_pointer,sync_count,avatar,city,geolocation FROM users WHERE password = ? AND uid = ?`, token, uid))
	if err == sql.ErrNoRows {
		return nil, "身份验证失败：查无此人"
	}
	if err != nil {
		return nil, "mysql: 获取身份信息错误"
	}
	return user, ""
}

func ScanUser(row interface {
	Scan(dest ...any) error
}) (*models.User, error) {
	u := &models.User{}
	var reg, last, comment, wx, phone, homepage, gaode, avatar, city, geo sql.NullString
	var countMapPtr sql.NullInt64
	err := row.Scan(
		&u.UID, &u.Email, &u.Nickname, &u.Username, &u.Password,
		&reg, &last, &comment, &wx, &phone, &homepage, &gaode, &u.GroupID,
		&u.CountDiary, &u.CountDict, &u.CountQR, &u.CountWords, &u.CountMapRoute,
		&countMapPtr, &u.SyncCount, &avatar, &city, &geo,
	)
	if err != nil {
		return nil, err
	}
	if reg.Valid {
		u.RegisterTime = &reg.String
	}
	if last.Valid {
		u.LastVisitTime = &last.String
	}
	if comment.Valid {
		u.Comment = &comment.String
	}
	if wx.Valid {
		u.Wx = &wx.String
	}
	if phone.Valid {
		u.Phone = &phone.String
	}
	if homepage.Valid {
		u.Homepage = &homepage.String
	}
	if gaode.Valid {
		u.Gaode = &gaode.String
	}
	if countMapPtr.Valid {
		v := int(countMapPtr.Int64)
		u.CountMapPointer = &v
	}
	if avatar.Valid {
		u.Avatar = &avatar.String
	}
	if city.Valid {
		u.City = &city.String
	}
	if geo.Valid {
		u.Geolocation = &geo.String
	}
	return u, nil
}

func RequireAdmin(c *gin.Context) bool {
	u := GetUser(c)
	return u != nil && u.IsAdmin()
}

package middleware

import (
	"database/sql"

	"github.com/KyleBing/portal-go/internal/auth"
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

// VerifyAuthorization 校验 Authorization: Bearer <jwt>，再按 uid 加载用户。
func VerifyAuthorization(c *gin.Context) (*models.User, string) {
	raw := auth.BearerToken(c.GetHeader("Authorization"))
	if raw == "" {
		return nil, "无 token"
	}
	claims, err := auth.Parse(raw)
	if err != nil {
		return nil, "身份验证失败：token 无效或已过期"
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		return nil, "mysql: 获取身份信息错误"
	}
	user, err := ScanUser(diary.QueryRow(`SELECT uid,email,nickname,username,password,register_time,last_visit_time,comment,wx,phone,homepage,gaode,group_id,count_diary,count_dict,count_qr,count_words,count_map_route,count_map_pointer,sync_count,avatar,city,geolocation FROM users WHERE uid = ?`, claims.UID))
	if err == sql.ErrNoRows {
		return nil, "身份验证失败：查无此人"
	}
	if err != nil {
		return nil, "mysql: 获取身份信息错误"
	}
	// 快到期时静默续签，通过响应头交给前端更新
	if auth.ShouldRenew(claims) {
		if newTok, err := auth.Issue(user); err == nil {
			c.Header(auth.HeaderRenewedToken, newTok)
		}
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

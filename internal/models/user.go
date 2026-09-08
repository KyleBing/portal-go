package models

type User struct {
	UID             int64   `json:"uid"`
	Email           string  `json:"email"`
	Nickname        string  `json:"nickname"`
	Username        string  `json:"username"`
	Password        string  `json:"password"`
	RegisterTime    *string `json:"register_time"`
	LastVisitTime   *string `json:"last_visit_time"`
	Comment         *string `json:"comment"`
	Wx              *string `json:"wx"`
	Phone           *string `json:"phone"`
	Homepage        *string `json:"homepage"`
	Gaode           *string `json:"gaode"`
	GroupID         int     `json:"group_id"`
	CountDiary      int     `json:"count_diary"`
	CountDict       int     `json:"count_dict"`
	CountQR         int     `json:"count_qr"`
	CountWords      int     `json:"count_words"`
	CountMapRoute   int     `json:"count_map_route"`
	CountMapPointer *int    `json:"count_map_pointer"`
	SyncCount       int     `json:"sync_count"`
	Avatar          *string `json:"avatar"`
	City            *string `json:"city"`
	Geolocation     *string `json:"geolocation"`
}

func (u *User) IsAdmin() bool {
	return u.GroupID == 1
}

type Pager struct {
	PageSize int   `json:"pageSize"`
	PageNo   int   `json:"pageNo"`
	Total    int64 `json:"total"`
}

type ListResult struct {
	List  interface{} `json:"list"`
	Pager Pager       `json:"pager"`
}

package user

import (
	"database/sql"
	"strings"

	"github.com/KyleBing/portal-go/internal/apihelper"
)

// resolveRegisterChannel 在注册写入 users 前判定来源。
func resolveRegisterChannel(diary *sql.DB, body map[string]interface{}, isFirstUser bool) string {
	if isFirstUser {
		return "bootstrap"
	}
	invitationCode := strings.TrimSpace(apihelper.S(body, "invitationCode"))
	if invitationCode == "" {
		return "unknown"
	}
	inv, err := apihelper.QueryMap(diary, `SELECT applicant_email FROM invitations WHERE id = ?`, invitationCode)
	if err != nil || inv == nil {
		return "config"
	}
	if email := strings.TrimSpace(apihelper.S(inv, "applicant_email")); email != "" {
		return "paid"
	}
	return "gift"
}

package middleware

const (
	CtxUserID   = "user_id"
	CtxAdminID  = "admin_id"
	CtxResponse = "response"

	CookieNameRefreshToken      = "refresh_token"
	CookieNameAdminRefreshToken = "admin_refresh_token"
	HeaderNameCSRFToken         = "X-CSRF-Token"
)

package models

// AdminUserListResponse represents the members visible in the administration area.
type AdminUserListResponse struct {
	TotalUserCount int64            `json:"totalUserCount,string"`
	Users          []*AdminUserInfo `json:"users"`
}

// AdminUserInfo is the limited user information exposed to an administrator.
type AdminUserInfo struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Nickname        string `json:"nickname"`
	Disabled        bool   `json:"disabled"`
	EmailVerified   bool   `json:"emailVerified"`
	CreatedUnixTime int64  `json:"createdAt"`
	LastLoginAt     int64  `json:"lastLoginAt"`
}

// AdminUserPasswordUpdateRequest represents a password reset requested by an administrator.
type AdminUserPasswordUpdateRequest struct {
	Username string `json:"username" binding:"required,notBlank,max=32,validUsername"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// AdminUserActionRequest represents an administration action for one user.
type AdminUserActionRequest struct {
	Username string `json:"username" binding:"required,notBlank,max=32,validUsername"`
}

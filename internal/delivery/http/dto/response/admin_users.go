package response

import "time"

type AdminUserResponse struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	ProfileImage    string     `json:"profile_image"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	Role            string     `json:"role"`
	IsBlocked       bool       `json:"is_blocked"`
	LastSeen        *time.Time `json:"last_seen"`
	CreatedAt       time.Time  `json:"created_at"`
}

type AdminUserListResponse struct {
	Users      []AdminUserResponse `json:"users"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	Total      int64               `json:"total"`
	TotalPages int                 `json:"total_pages"`
}


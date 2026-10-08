package response

import "time"

type ProfileResponse struct {
	ID              uint       `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Phone           string     `json:"phone"`
	ProfileImage    string     `json:"profile_image"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
}
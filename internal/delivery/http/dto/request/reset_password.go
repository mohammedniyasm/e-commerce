package request

type ResetPasswordRequest struct{
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
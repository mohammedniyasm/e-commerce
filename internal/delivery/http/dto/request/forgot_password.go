package request

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

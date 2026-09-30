package request

type SendVerificationOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

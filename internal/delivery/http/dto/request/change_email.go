package request

type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

type VerifyEmailChangeRequest struct {
	OTP string `json:"otp" binding:"required,len=6"`
}
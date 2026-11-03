package request

type GoogleLoginRequest struct {
	IDToken      string `json:"id_token" binding:"required"`
	ReferralCode string `json:"referral_code"`
}

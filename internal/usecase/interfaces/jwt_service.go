package interfaces

import "time"

type AccessClaims struct {
	UserID    string    `json:"sub"`
	Role      string    `json:"role"`
	JTI       string    `json:"jti"`
	ExpiresAt time.Time `json:"expires_at"`
}
type JWTService interface {
	GenerateAccessToken(userID uint, role string) (string, error)
	GenerateRefreshToken(userID uint) (string, string, error)
	ValidateAccessToken(tokenString string) (*AccessClaims, error)
	ValidateRefreshToken(tokenString string) (string, string, error)
}

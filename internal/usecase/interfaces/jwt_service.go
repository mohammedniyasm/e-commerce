package interfaces

type AccessClaims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
}
type JWTService interface {
	GenerateAccessToken(userID uint, role string) (string, error)
	GenerateRefreshToken(userID uint) (string, string, error)
	ValidateAccessToken(tokenString string) (*AccessClaims, error)
	ValidateRefreshToken(tokenString string) (string, string, error)
}

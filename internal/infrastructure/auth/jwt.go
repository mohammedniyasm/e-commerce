package authinfra

import (
	"ecommerce/config"
	"ecommerce/internal/usecase/interfaces"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	config config.JWTConfig
}

func NewJWTService(config config.JWTConfig) *JWTService {
	return &JWTService{
		config: config,
	}
}

type AccessClaims struct {
	// UserId string `json:"sub"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}
type RefreshClaims struct {
	// UserId string `json:"sub"`
	jwt.RegisteredClaims
}

func (s *JWTService) GenerateAccessToken(userID uint, role string) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		// Subject: strconv.FormatUint(uint64(userID), 10),
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessExpiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.config.AccessSecret))
}
func (s *JWTService) GenerateRefreshToken(userID uint) (string, string, error) {
	now := time.Now()
	jti := uuid.NewString()
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshExpiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.config.RefreshSecret))
	if err != nil {
		return "", "", err
	}
	return signedToken, jti, nil
}
func (s *JWTService) ValidateAccessToken(tokenString string) (*interfaces.AccessClaims, error) {
	var claims AccessClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method,%s", token.Method.Alg())
			}
			return []byte(s.config.AccessSecret), nil
		})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("Invalid Access token")
	}
	return &interfaces.AccessClaims{
		UserID: claims.Subject,
		Role:   claims.Role,
	}, nil
}
func (s *JWTService) ValidateRefreshToken(tokenString string) (string, string, error) {
	var claims RefreshClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method, %s", t.Method.Alg())
		}
		return []byte(s.config.RefreshSecret), nil
	})
	if err != nil {
		return "", "", err
	}
	if !token.Valid{
		return "", "", fmt.Errorf("invalid refresh token")
	}
	return claims.Subject, claims.ID, nil
}

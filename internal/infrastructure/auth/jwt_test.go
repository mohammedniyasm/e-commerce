package authinfra

import (
	"ecommerce/config"
	"testing"
	"time"
)
func TestJWTService_ValidateRefreshToken_Expired(t *testing.T){
		config := config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  -1 * time.Minute,
		RefreshExpiry: -time.Hour,
	}
	service := NewJWTService(config)
	token,_,err:=service.GenerateRefreshToken(1)
	if err!=nil{
		t.Fatalf("failed to generate refresh token: %s",err)
	}
	_,_,err=service.ValidateRefreshToken(token)
	if err == nil{
		t.Fatalf("expected error for expired refresh token")
	}


}
func TestJWTService_GenerateAndValidateRefreshToken(t *testing.T) {
	config := config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  -1 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
	service := NewJWTService(config)

	token, expectedJTI, err := service.GenerateRefreshToken(1)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	userID, jti, err := service.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("failed to validate refresh token: %v", err)
	}

	if userID != "1" {
		t.Fatalf("expected user ID 1, got %s", userID)
	}

	if jti != expectedJTI {
		t.Fatalf("expected JTI %s, got %s", expectedJTI, jti)
	}
}
func TestJWTService_ValidateRefreshToken_Tampered(t *testing.T) {
	config := config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  -1 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
	service := NewJWTService(config)

	token, _, err := service.GenerateRefreshToken(1)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	// Modify the token
	tamperedToken := token[:len(token)-1] + "x"

	_, _, err = service.ValidateRefreshToken(tamperedToken)
	if err == nil {
		t.Fatal("expected error for tampered refresh token")
	}
}

// func TestJWTService_RejectsExpiredAccessToken(t *testing.T) {
// 	cfg := config.JWTConfig{
// 		AccessSecret:  "test-access-secret",
// 		RefreshSecret: "test-refresh-secret",
// 		AccessExpiry:  -1 * time.Minute,
// 		RefreshExpiry: 7 * 24 * time.Hour,
// 	}

// 	service := NewJWTService(cfg)

// 	token, err := service.GenerateAccessToken(123, "user")
// 	if err != nil {
// 		t.Fatalf("failed to generate access token: %v", err)
// 	}

// 	_, err = service.ValidateAccessToken(token)

// 	if err == nil {
// 		t.Fatal("expected expired token to be rejected")
// 	}
// }
// func TestJWTService_GenerateAndValidateAccessToken(t *testing.T) {
// 	cfg := config.JWTConfig{
// 		AccessSecret:  "Hello",
// 		RefreshSecret: "World",
// 		AccessExpiry:  15 * time.Minute,
// 		RefreshExpiry: 7 * 24 * time.Hour,
// 	}
// 	service := NewJWTService(cfg)
// 	token, err := service.GenerateAccessToken(123, "user")
// 	if err != nil {
// 		t.Fatalf("failed to generate access token: %v", err)
// 	}
// 	claims, err := service.ValidateAccessToken(token)
// 	if err != nil {
// 		t.Fatalf("failed to validate accesss token: %v", err)
// 	}
// 	if claims.UserID != "123" {
// 		t.Errorf("expected user ID 123,got %s", claims.UserID)
// 	}
// 	if claims.Role != "user" {
// 		t.Errorf("expected role user ,got %s", claims.Role)
// 	}

// }
// func TestJWTService_RejectsTamperedAccessToken(t *testing.T) {
// 	cfg := config.JWTConfig{
// 		AccessSecret:  "test-access-secret",
// 		RefreshSecret: "test-refresh-secret",
// 		AccessExpiry:  15 * time.Minute,
// 		RefreshExpiry: 7 * 24 * time.Hour,
// 	}

// 	service := NewJWTService(cfg)

// 	token, err := service.GenerateAccessToken(123, "user")
// 	if err != nil {
// 		t.Fatalf("failed to generate access token: %v", err)
// 	}

// 	// Tamper with the token
// 	tamperedToken := token[:len(token)-1] + "x"

// 	_, err = service.ValidateAccessToken(tamperedToken)

// 	if err == nil {
// 		t.Fatal("expected tampered token to be rejected")
// 	}
// }

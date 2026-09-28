package authinfra

import (
	"ecommerce/config"
	"testing"
	"time"
)

func TestJWTService_RejectsExpiredAccessToken(t *testing.T) {
	cfg := config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  -1 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service := NewJWTService(cfg)

	token, err := service.GenerateAccessToken(123, "user")
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	_, err = service.ValidateAccessToken(token)

	if err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}
func TestJWTService_GenerateAndValidateAccessToken(t *testing.T) {
	cfg := config.JWTConfig{
		AccessSecret:  "Hello",
		RefreshSecret: "World",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
	service := NewJWTService(cfg)
	token, err := service.GenerateAccessToken(123, "user")
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}
	claims, err := service.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("failed to validate accesss token: %v", err)
	}
	if claims.UserID != "123" {
		t.Errorf("expected user ID 123,got %s", claims.UserID)
	}
	if claims.Role != "user" {
		t.Errorf("expected role user ,got %s", claims.Role)
	}

}
func TestJWTService_RejectsTamperedAccessToken(t *testing.T) {
	cfg := config.JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}

	service := NewJWTService(cfg)

	token, err := service.GenerateAccessToken(123, "user")
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	// Tamper with the token
	tamperedToken := token[:len(token)-1] + "x"

	_, err = service.ValidateAccessToken(tamperedToken)

	if err == nil {
		t.Fatal("expected tampered token to be rejected")
	}
}

package authinfra

import (
	"context"
	"ecommerce/internal/usecase/interfaces"
	"fmt"
	"strings"

	"google.golang.org/api/idtoken"
)

type GoogleTokenVerifier struct {
	clientID string
}

func NewGoogleTokenVerifier(clientID string) *GoogleTokenVerifier {
	return &GoogleTokenVerifier{
		clientID: clientID,
	}
}
func (v *GoogleTokenVerifier) Verify(ctx context.Context, rawIDToken string) (*interfaces.GoogleUserInfo, error) {
	if strings.TrimSpace(rawIDToken) == "" {
		return nil, fmt.Errorf("google id token is required")
	}
	payload, err := idtoken.Validate(ctx, rawIDToken, v.clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid google id token: %w", err)
	}
	if payload.Subject == "" {
		return nil, fmt.Errorf("google subject is missing")
	}
	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("google email is missing")
	}
	emailVerified, ok := payload.Claims["email_verified"].(bool)
	if !ok || !emailVerified {
		return nil, fmt.Errorf("google email is not verified")
	}
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)
	return &interfaces.GoogleUserInfo{
		SubjectID: payload.Subject,
		Email:     email,
		Name:      name,
		Picture:   picture,
	}, nil
}

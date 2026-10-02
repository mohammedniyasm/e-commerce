package interfaces

import "context"

type GoogleUserInfo struct {
	SubjectID string
	Email     string
	Name      string
	Picture   string
}
type GoogleTokenVerifier interface {
	Verify(ctx context.Context,idToken string)(*GoogleUserInfo,error)
}
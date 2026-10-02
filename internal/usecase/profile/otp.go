package profile

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const emailChangeOTPExpiry = 5 * time.Minute

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(100000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
func emailChangeOTPKey(userID uint) string {
	return fmt.Sprintf("email_change_otp:%d", userID)
}
func emailChangePendingEmailKey(userID uint) string {
	return fmt.Sprintf("email_change_pending_email:%d", userID)
}

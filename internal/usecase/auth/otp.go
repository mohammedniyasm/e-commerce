package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)
const emailVerificationOTPExpiry=5*time.Minute
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader,big.NewInt(100000))
	if err != nil{
		return "",err
	}
	return fmt.Sprintf("%06d",n.Int64()), nil
}
func emailVerificationOTPKey(email string)string{
	return "email_verification_otp:"+email
}
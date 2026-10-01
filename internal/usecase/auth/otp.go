package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)
const emailVerificationOTPExpiry=5*time.Minute
const passwordResetOTPExpiry=5*time.Minute
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
func passwordResetOTPKey(email string)string{
	return "password_reset_otp:"+email
}
func generaratePasswordResetToken()(string,error){
	b:=make([]byte,32)
	if _,err:=rand.Read(b);err!=nil{
		return "",err
	}
	return fmt.Sprintf("%x",b),nil
}
const passwordResetTokenExpiry = 10*time.Minute
func passwordResetTokenKey(token string)string{
	return "password_reset_token:"+token
}
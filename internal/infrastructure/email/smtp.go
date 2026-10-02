package email

import (
	"context"
	"ecommerce/config"
	"fmt"
	"net/smtp"
)

type SMTPEmailSender struct {
	config config.EmailConfig
}

func NewSMTPEmailSender(config config.EmailConfig) *SMTPEmailSender {
	return &SMTPEmailSender{
		config: config,
	}
}
func (s *SMTPEmailSender) SendVerificationOTP(ctx context.Context, to string, otp string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)
	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: Verify your email\r\n"+
		"Content-Type:text/plain;charset=UTF-8\r\n"+
		"\r\n"+
		"Your email verification OTP is: %s\r\n"+
		"This OTP will expire in 5 minutes.\r\n", s.config.From, to, otp,
	)
	address:=fmt.Sprintf("%s:%d",s.config.SMTPHost,s.config.SMTPPort)
	return smtp.SendMail(address,auth,s.config.From,[]string{to},[]byte(message))
}

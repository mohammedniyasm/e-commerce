package validator

import (
	"regexp"
	"strings"
	"unicode"
)

func ValidateName(name string) bool {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 30 {
		return false
	}
	for _, r := range name {
		if unicode.IsLetter(r) || r == ' ' || r == '-' || r == '\'' {
			continue
		}
		return false
	}
	return true
}
func ValidatePhone(phone string) bool {
	matched, _ := regexp.MatchString(`^[0-9]{10}$`, phone)
	return matched
}
func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	return hasLower && hasUpper && hasNumber && hasSpecial
}

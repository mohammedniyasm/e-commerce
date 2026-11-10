package validator

import (
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		test string
		name string
		want bool
	}{
		{
			test: "Valid name", name: "Niyas", want: true,
		},
		{
			test: "full space", name: "    ", want: false,
		},
		{
			test: "invalid character", name: "#Niyas", want: false,
		},
		{
			test: "invalid digits", name: "122NiYas", want: false,
		},
		{
			test: "unwanted digits", name: "Niyas 123@3 Muthu", want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			got := ValidateName(tt.name)
			if got != tt.want {
				t.Errorf("Caught;name: %v, got: %v,want: %v", tt.name, got, tt.want)
			}
		})
	}
	t.Log("Test completed")
}
func TestValidatePhone(t *testing.T) {
	tests := []struct {
		test  string
		phone string
		want  bool
	}{
		{
			test: "Valid phone", phone: "8848598156", want: true,
		},
		{
			test: "full space", phone: "    ", want: false,
		},
		{
			test: "invalid character", phone: "#884859815", want: false,
		},
		{
			test: "invalid characters", phone: "122NiYas", want: false,
		},
		{
			test: "Morethan count", phone: "123123123123123", want: false,
		},
		{
			test: "Less than count", phone: "123123", want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			got := ValidatePhone(tt.phone)
			if got != tt.want {
				t.Errorf("Caught;phone: %v, got: %v,want: %v,len; %v", tt.phone, got, tt.want, len(tt.phone))
			}
		})
	}
	t.Log("Test completed")
}
func TestValidatePassword(t *testing.T) {
	tests := []struct {
		test     string
		password string
		want     bool
	}{
		{
			test: "Valid Password", password: "Niyas@123", want: true,
		},
		{
			test: "full space", password: "    ", want: false,
		},
		{
			test: "small letters only", password: "niiyaaaaas", want: false,
		},
		{
			test: "capital letters only", password: "NIIYAAAAS", want: false,
		},
		{
			test: "characters only", password: "Niyaaaaaaas", want: false,
		},
		{
			test: "digits only", password: "84884848484", want: false,
		},
		{
			test: "symbols only", password: "@##@#$%#$%@", want: false,
		},
		{
			test: "Less Count", password: "Ani@12", want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			got := ValidatePassword(tt.password)
			if got != tt.want {
				t.Errorf("Caught;password: %v, got: %v,want: %v", tt.password, got, tt.want)
			}
		})
	}
	t.Log("Test completed")
}

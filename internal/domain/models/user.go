package models

import "time"

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID              uint   `gorm:"primaryKey"`
	Name            string `gorm:"not null"`
	Email           string `gorm:"unique;not null"`
	Phone           string `gorm:"size:20"`
	Password        *string
	EmailVerifiedAt *time.Time
	ProfileImage    string   `gorm:"size:255"`
	Role            UserRole `gorm:"not null;default:user"`
	IsBlocked       bool     `gorm:"not null;default:false"`
	LastSeen        *time.Time
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}

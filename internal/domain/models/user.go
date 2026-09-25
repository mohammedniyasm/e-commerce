package models

import "time"

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID        uint
	Name      string
	Email     string
	Phone     string
	Password  *string
	Role      UserRole
	IsBlocked bool
	LastSeen  *time.Time
	CreatedAt time.Time
	UpdatedAt *time.Time
}

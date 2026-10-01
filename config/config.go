package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Cookie   CookieConfig
	Email    EmailConfig
	Google   GoogleConfig
}
type ServerConfig struct {
	Port string
}
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}
type RedisConfig struct {
	Host     string
	Port     string
	Password string
}
type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}
type CookieConfig struct {
	RefreshTokenName string
	HTTPOnly         bool
	Secure           bool
	SameSite         string
	Path             string
	MaxAge           int

	PasswordResetTokenName string
	PasswordResetHTTPOnly  bool
	PasswordResetSecure    bool
	PasswordResetSameSite  string
	PasswordResetPath      string
}
type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
}
type GoogleConfig struct {
	ClientID string
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()
	accessExpiry, err := time.ParseDuration(os.Getenv("JWT_ACCESS_EXPIRY"))
	if err != nil {
		return Config{}, err
	}
	refreshExpiry, err := time.ParseDuration(os.Getenv("JWT_REFRESH_EXPIRY"))
	if err != nil {
		return Config{}, err
	}
	cookieHTTPOnly, err := strconv.ParseBool(
		os.Getenv("REFRESH_TOKEN_COOKIE_HTTP_ONLY"),
	)
	if err != nil {
		return Config{}, err
	}

	cookieSecure, err := strconv.ParseBool(
		os.Getenv("REFRESH_TOKEN_COOKIE_SECURE"),
	)
	if err != nil {
		return Config{}, err
	}
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return Config{}, err
	}
	passwordResetHTTPOnly, err := strconv.ParseBool(
		os.Getenv("PASSWORD_RESET_COOKIE_HTTP_ONLY"),
	)
	if err != nil {
		return Config{}, err
	}

	passwordResetSecure, err := strconv.ParseBool(
		os.Getenv("PASSWORD_RESET_COOKIE_SECURE"),
	)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Server: ServerConfig{
			Port: os.Getenv("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		},
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
		},
		JWT: JWTConfig{
			AccessSecret:  os.Getenv("JWT_ACCESS_SECRET"),
			RefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		Cookie: CookieConfig{
			RefreshTokenName: os.Getenv("REFRESH_TOKEN_COOKIE_NAME"),
			HTTPOnly:         cookieHTTPOnly,
			Secure:           cookieSecure,
			SameSite:         os.Getenv("REFRESH_TOKEN_COOKIE_SAME_SITE"),
			Path:             os.Getenv("REFRESH_TOKEN_COOKIE_PATH"),
			MaxAge:           int(refreshExpiry.Seconds()),

			PasswordResetTokenName: os.Getenv("PASSWORD_RESET_COOKIE_NAME"),
			PasswordResetHTTPOnly:  passwordResetHTTPOnly,
			PasswordResetSecure:    passwordResetSecure,
			PasswordResetSameSite:  os.Getenv("PASSWORD_RESET_COOKIE_SAME_SITE"),
			PasswordResetPath:      os.Getenv("PASSWORD_RESET_COOKIE_PATH"),
		},
		Email: EmailConfig{
			SMTPHost: os.Getenv("SMTP_HOST"),
			SMTPPort: smtpPort,
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
		},
		Google: GoogleConfig{
			ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
		},
	}, nil
}

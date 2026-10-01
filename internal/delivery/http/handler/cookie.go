package handler

import (
	"ecommerce/config"
	"net/http"
)

func SetRefreshTokenCookie(w http.ResponseWriter, cookieConfig config.CookieConfig, refreshToken string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:  cookieConfig.RefreshTokenName,
		Value: refreshToken,
		// Path: cookieConfig.Path,
		HttpOnly: cookieConfig.HTTPOnly,
		Secure:   cookieConfig.Secure,
		SameSite: getSameSiteMode(cookieConfig.SameSite),
		MaxAge:   maxAge,
	})
}
func clearRefreshTokenCookie(w http.ResponseWriter, cookieConfig config.CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:  cookieConfig.RefreshTokenName,
		Value: "",
		// Path:     cookieConfig.Path,
		HttpOnly: cookieConfig.HTTPOnly,
		Secure:   cookieConfig.Secure,
		SameSite: getSameSiteMode(cookieConfig.SameSite),
		MaxAge:   -1,
	})
}
func getSameSiteMode(value string) http.SameSite {
	switch value {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

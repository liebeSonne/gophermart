package handler

import "net/http"

type CookieService interface {
	SetAuthToken(tokenString string, w http.ResponseWriter, r *http.Request) error
}

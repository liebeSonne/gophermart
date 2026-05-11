package cookie

import (
	"errors"
	"net/http"
)

func NewService(
	tokenKey string,
) *Service {
	return &Service{
		tokenKey: tokenKey,
	}
}

type Service struct {
	tokenKey string
}

func (s *Service) SetAuthToken(tokenString string, w http.ResponseWriter, r *http.Request) error {
	cookie := &http.Cookie{
		Name:  s.tokenKey,
		Value: tokenString,
	}
	http.SetCookie(w, cookie)
	r.AddCookie(cookie)
	return nil
}

func (s *Service) GetAuthToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(s.tokenKey)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", nil
		}
		return "", err
	}

	if cookie == nil || cookie.Value == "" {
		return "", nil
	}

	return cookie.Value, nil
}

package handler

import "github.com/liebeSonne/gophermart/internal/auth"

type TokenService interface {
	Create(tokenData auth.Token) (string, error)
}

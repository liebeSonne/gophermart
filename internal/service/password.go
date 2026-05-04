package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type PasswordService interface {
	CreateHash(ctx context.Context, password string) (string, error)
	CheckHash(ctx context.Context, password, hash string) (bool, error)
}

func NewPasswordService(
	secretKey []byte,
) PasswordService {
	return &passwordService{
		secretKey: secretKey,
	}
}

type passwordService struct {
	secretKey []byte
}

func (s *passwordService) CreateHash(_ context.Context, password string) (string, error) {
	return s.createHash(password), nil
}

func (s *passwordService) CheckHash(_ context.Context, password, hash string) (bool, error) {
	passHash := s.createHash(password)

	return hash == passHash, nil
}

func (s *passwordService) createHash(password string) string {
	template := s.createTemplate(password)

	h := sha256.New()
	h.Write(template)
	passHash := h.Sum(nil)

	return hex.EncodeToString(passHash)
}

func (s *passwordService) createTemplate(password string) []byte {
	template := fmt.Sprintf("p_%s_s_%c", password, s.secretKey)

	return []byte(template)
}

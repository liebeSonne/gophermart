package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/liebeSonne/gophermart/internal/model"
	"github.com/liebeSonne/gophermart/internal/service"
)

func NewUserRepository(
	client ContextClient,
) *UserRepository {
	return &UserRepository{
		client: client,
	}
}

type UserRepository struct {
	client ContextClient
}

func (r *UserRepository) NextID(_ context.Context) uuid.UUID {
	return uuid.New()
}

func (r *UserRepository) Store(ctx context.Context, user model.User) error {
	const sqlQuery = `
		INSERT INTO "user" (id, login, passhash) VALUES ($1, $2, $3)
	`

	_, err := r.client.Exec(ctx, sqlQuery, user.ID, user.Login, user.PassHash)
	if err != nil {
		return fmt.Errorf("error on insert user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID uuid.UUID) (model.User, error) {
	const sqlQuery = `
		SELECT u.id, u.login, u.passhash 
		FROM "user" u
		WHERE u.id = $1 
		LIMIT 1
	`

	var user model.User
	row := r.client.QueryRow(ctx, sqlQuery, userID)
	err := row.Scan(&user.ID, &user.Login, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, fmt.Errorf("user %s not found: %w", userID.String(), service.ErrUserNotFound)
		}
		return model.User{}, fmt.Errorf("error on scan row: %w", err)
	}

	return user, nil
}

func (r *UserRepository) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	const sqlQuery = `
		SELECT u.id, u.login, u.passhash 
		FROM "user" u
		WHERE u.login = $1 
		LIMIT 1
	`

	var user model.User
	row := r.client.QueryRow(ctx, sqlQuery, login)
	err := row.Scan(&user.ID, &user.Login, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error on scan row: %w", err)
	}

	return &user, nil
}

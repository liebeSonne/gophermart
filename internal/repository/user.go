package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/liebeSonne/gophermart/internal/model"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	NextID(ctx context.Context) uuid.UUID
	Store(ctx context.Context, user model.User) error
	GetByID(ctx context.Context, userID uuid.UUID) (model.User, error)
	FindByLogin(ctx context.Context, login string) (*model.User, error)
}

func NewUserRepository(
	pool *pgxpool.Pool,
) UserRepository {
	return &userRepository{
		pool: pool,
	}
}

type userRepository struct {
	pool *pgxpool.Pool
}

func (r *userRepository) NextID(_ context.Context) uuid.UUID {
	return uuid.New()
}

func (r *userRepository) Store(ctx context.Context, user model.User) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("error on begin transaction: %w", err)
	}

	defer func() {
		err = tx.Rollback(ctx)
		if err != nil {
			fmt.Printf("error on rollback transaction: %v\n", err)
		}
	}()

	const sqlQuery = `
		INSERT INTO "user" (id, login, passhash) VALUES ($1, $2, $3)
	`

	_, err = tx.Exec(ctx, sqlQuery, user.ID, user.Login, user.PassHash)
	if err != nil {
		return fmt.Errorf("error on insert user: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error on commit transaction: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (model.User, error) {
	const sqlQuery = `
		SELECT u.id, u.login, u.passhash 
		FROM "user" u
		WHERE u.id = $1 
		LIMIT 1
	`

	var user model.User
	row := r.pool.QueryRow(ctx, sqlQuery, userID)
	err := row.Scan(&user.ID, &user.Login, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, fmt.Errorf("user %s not found: %w", userID.String(), ErrUserNotFound)
		}
		return model.User{}, fmt.Errorf("error on scan row: %w", err)
	}

	return user, nil
}

func (r *userRepository) FindByLogin(ctx context.Context, login string) (*model.User, error) {
	const sqlQuery = `
		SELECT u.id, u.login, u.passhash 
		FROM "user" u
		WHERE u.login = $1 
		LIMIT 1
	`

	var user model.User
	row := r.pool.QueryRow(ctx, sqlQuery, login)
	err := row.Scan(&user.ID, &user.Login, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error on scan row: %w", err)
	}

	return &user, nil
}

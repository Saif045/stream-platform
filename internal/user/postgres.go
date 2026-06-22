package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}

var _ Store = (*PostgresStore)(nil)

func (s *PostgresStore) Create(ctx context.Context, user *User) error {
	err := s.db.QueryRow(
		ctx,
		`
		INSERT INTO users (
			id,
			username,
			password_hash
		)
		VALUES ($1, $2, $3)
		RETURNING created_at
		`,
		user.ID,
		user.Username,
		user.PasswordHash,
	).Scan(&user.CreatedAt)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *PostgresStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}

	err := s.db.QueryRow(
		ctx,
		`
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE username = $1
		`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}

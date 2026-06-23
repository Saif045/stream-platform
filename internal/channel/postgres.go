package channel

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}

var _ Store = (*PostgresStore)(nil)

func (s *PostgresStore) Create(ctx context.Context, channel *Channel) error {
	err := s.db.QueryRow(
		ctx,
		`
		INSERT INTO channels (
			id,
			user_id,
			slug
		)
		VALUES ($1, $2, $3)
		RETURNING created_at
		`,
		channel.ID,
		channel.UserID,
		channel.Slug,
	).Scan(&channel.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrSlugTaken
		}

		return fmt.Errorf("create channel: %w", err)
	}

	return nil
}

func (s *PostgresStore) GetByID(ctx context.Context, id string) (*Channel, error) {
	channel := &Channel{}

	err := s.db.QueryRow(
		ctx,
		`
		SELECT id, user_id, slug, created_at
		FROM channels
		WHERE id = $1
		`,
		id,
	).Scan(
		&channel.ID,
		&channel.UserID,
		&channel.Slug,
		&channel.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("get channel: %w", err)
	}

	return channel, nil
}

func (s *PostgresStore) List(ctx context.Context) ([]*Channel, error) {
	rows, err := s.db.Query(
		ctx,
		`
		SELECT id, user_id, slug, created_at
		FROM channels
		ORDER BY created_at DESC
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer rows.Close()

	channels := make([]*Channel, 0)

	for rows.Next() {
		channel := &Channel{}

		if err := rows.Scan(
			&channel.ID,
			&channel.UserID,
			&channel.Slug,
			&channel.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}

		channels = append(channels, channel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate channels: %w", err)
	}

	return channels, nil
}

func (s *PostgresStore) GetBySlug(ctx context.Context, slug string) (*Channel, error) {
	channel := &Channel{}

	err := s.db.QueryRow(
		ctx,
		`
		SELECT id, user_id, slug, created_at
		FROM channels
		WHERE slug = $1
		`,
		slug,
	).Scan(
		&channel.ID,
		&channel.UserID,
		&channel.Slug,
		&channel.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("get channel by slug: %w", err)
	}

	return channel, nil
}

package channel

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

var _ Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) Create(channel *Channel) error {
	err := r.db.QueryRow(
		context.Background(),
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
		return fmt.Errorf("create channel: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(id string) (*Channel, error) {
	channel := &Channel{}

	err := r.db.QueryRow(
		context.Background(),
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
		return nil, fmt.Errorf("get channel: %w", err)
	}

	return channel, nil
}

func (r *PostgresRepository) List() ([]*Channel, error) {
	rows, err := r.db.Query(
		context.Background(),
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
func (r *PostgresRepository) GetBySlug(slug string) (*Channel, error) {
	channel := &Channel{}

	err := r.db.QueryRow(
		context.Background(),
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
		return nil, fmt.Errorf("get channel by slug: %w", err)
	}

	return channel, nil
}

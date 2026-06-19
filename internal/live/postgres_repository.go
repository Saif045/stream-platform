package live

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

func (r *PostgresRepository) Create(stream *Stream) error {
	err := r.db.QueryRow(
		context.Background(),
		`
		INSERT INTO streams (
			id,
			stream_key,
			status,
			error
		)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
		`,
		stream.ID,
		stream.StreamKey,
		stream.Status,
		nilIfEmpty(stream.Error),
	).Scan(&stream.CreatedAt)

	if err != nil {
		return fmt.Errorf("create stream: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(id string) (*Stream, error) {
	return r.getOne(
		`
		SELECT
			id,
			stream_key,
			status,
			COALESCE(error, ''),
			created_at,
			started_at,
			stopped_at
		FROM streams
		WHERE id = $1
		`,
		id,
	)
}

func (r *PostgresRepository) GetByStreamKey(streamKey string) (*Stream, error) {
	return r.getOne(
		`
		SELECT
			id,
			stream_key,
			status,
			COALESCE(error, ''),
			created_at,
			started_at,
			stopped_at
		FROM streams
		WHERE stream_key = $1
		`,
		streamKey,
	)
}

func (r *PostgresRepository) List() ([]*Stream, error) {
	rows, err := r.db.Query(
		context.Background(),
		`
		SELECT
			id,
			stream_key,
			status,
			COALESCE(error, ''),
			created_at,
			started_at,
			stopped_at
		FROM streams
		ORDER BY created_at DESC
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("list streams: %w", err)
	}
	defer rows.Close()

	streams := make([]*Stream, 0)

	for rows.Next() {
		stream := &Stream{}

		if err := rows.Scan(
			&stream.ID,
			&stream.StreamKey,
			&stream.Status,
			&stream.Error,
			&stream.CreatedAt,
			&stream.StartedAt,
			&stream.StoppedAt,
		); err != nil {
			return nil, fmt.Errorf("scan stream: %w", err)
		}

		streams = append(streams, stream)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate streams: %w", err)
	}

	return streams, nil
}

func (r *PostgresRepository) Update(stream *Stream) error {
	result, err := r.db.Exec(
		context.Background(),
		`
		UPDATE streams
		SET
			status = $2,
			error = $3,
			started_at = $4,
			stopped_at = $5
		WHERE id = $1
		`,
		stream.ID,
		stream.Status,
		nilIfEmpty(stream.Error),
		stream.StartedAt,
		stream.StoppedAt,
	)
	if err != nil {
		return fmt.Errorf("update stream: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("live stream not found: %s", stream.ID)
	}

	return nil
}

func (r *PostgresRepository) getOne(query string, arg string) (*Stream, error) {
	stream := &Stream{}

	err := r.db.QueryRow(context.Background(), query, arg).Scan(
		&stream.ID,
		&stream.StreamKey,
		&stream.Status,
		&stream.Error,
		&stream.CreatedAt,
		&stream.StartedAt,
		&stream.StoppedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get stream: %w", err)
	}

	return stream, nil
}

func nilIfEmpty(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

package live

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

func (s *PostgresStore) Create(ctx context.Context, stream *Stream) error {
	err := s.db.QueryRow(
		ctx,
		`
		INSERT INTO streams (
			id,
			channel_id,
			stream_key,
			status,
			error
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
		`,
		stream.ID,
		stream.ChannelID,
		stream.StreamKey,
		stream.Status,
		nilIfEmpty(stream.Error),
	).Scan(&stream.CreatedAt)

	if err != nil {
		return fmt.Errorf("create stream: %w", err)
	}

	return nil
}

func (s *PostgresStore) GetByID(ctx context.Context, id string) (*Stream, error) {
	return s.getOne(
		ctx,
		`
		SELECT
			id,
			channel_id,
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

func (s *PostgresStore) GetByStreamKey(ctx context.Context, streamKey string) (*Stream, error) {
	return s.getOne(
		ctx,
		`
		SELECT
			id,
			channel_id,
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

func (s *PostgresStore) List(ctx context.Context) ([]*Stream, error) {
	rows, err := s.db.Query(
		ctx,
		`
		SELECT
			id,
			channel_id,
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
			&stream.ChannelID,
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

func (s *PostgresStore) Update(ctx context.Context, stream *Stream) error {
	result, err := s.db.Exec(
		ctx,
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

func (s *PostgresStore) ListByChannelID(ctx context.Context, channelID string) ([]*Stream, error) {
	rows, err := s.db.Query(
		ctx,
		`
		SELECT
			id,
			channel_id,
			stream_key,
			status,
			COALESCE(error, ''),
			created_at,
			started_at,
			stopped_at
		FROM streams
		WHERE channel_id = $1
		ORDER BY created_at DESC
		`,
		channelID,
	)
	if err != nil {
		return nil, fmt.Errorf("list streams by channel: %w", err)
	}
	defer rows.Close()

	streams := make([]*Stream, 0)

	for rows.Next() {
		stream := &Stream{}

		if err := rows.Scan(
			&stream.ID,
			&stream.ChannelID,
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

func (s *PostgresStore) GetLatestByChannelID(ctx context.Context, channelID string) (*Stream, error) {
	return s.getOne(
		ctx,
		`
		SELECT
			id,
			channel_id,
			stream_key,
			status,
			COALESCE(error, ''),
			created_at,
			started_at,
			stopped_at
		FROM streams
		WHERE channel_id = $1
		ORDER BY created_at DESC
		LIMIT 1
		`,
		channelID,
	)
}

func (s *PostgresStore) getOne(ctx context.Context, query string, arg string) (*Stream, error) {
	stream := &Stream{}

	err := s.db.QueryRow(ctx, query, arg).Scan(
		&stream.ID,
		&stream.ChannelID,
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

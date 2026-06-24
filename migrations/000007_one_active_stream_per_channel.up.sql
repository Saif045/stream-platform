CREATE UNIQUE INDEX streams_one_active_per_channel_idx
ON streams (channel_id)
WHERE status IN ('created', 'running');
ALTER TABLE streams
ADD COLUMN channel_id TEXT
REFERENCES channels(id);
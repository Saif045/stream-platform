.PHONY: server mediamtx publish create start

STREAM_ID ?= demo-$(shell date +%s)
CHANNEL_ID ?= channel-saif



server:
	go run ./cmd/server

mediamtx:
	mediamtx

create:
	curl -s -X POST http://localhost:8080/api/live/streams/create \
	-H "Content-Type: application/json" \
	-d '{"id":"$(STREAM_ID)","channel_id":"$(CHANNEL_ID)"}' | tee .stream.json

publish:
	ffmpeg -re -stream_loop -1 -i test.mp4 \
	-c copy \
	-f flv \
	$$(jq -r .rtmp_url .stream.json)

start:
	curl -X POST http://localhost:8080/api/live/streams/start \
	-H "Content-Type: application/json" \
	-d '{"id":"$(STREAM_ID)"}'


migrate-up:
	. ./.env && migrate -path migrations -database "$$MIGRATION_DATABASE_URL" up

migrate-version:
	. ./.env && migrate -path migrations -database "$$MIGRATION_DATABASE_URL" version

migrate-down:
	. ./.env && migrate -path migrations -database "$$MIGRATION_DATABASE_URL" down 1
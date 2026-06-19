.PHONY: server mediamtx publish create start

STREAM_ID ?= demo-$(shell date +%s)

server:
	go run ./cmd/server

mediamtx:
	mediamtx

create:
	curl -s -X POST http://localhost:8080/api/live/streams/create \
	-H "Content-Type: application/json" \
	-d '{"id":"$(STREAM_ID)"}' | tee .stream.json

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
	export $$(grep -v '^#' .env | xargs) && \
	migrate -path migrations -database "$$DATABASE_URL" up

# migrate-down:
# 	export $$(grep -v '^#' .env | xargs) && \
# 	migrate -path migrations -database "$$DATABASE_URL" down 1

migrate-version:
 	export $$(grep -v '^#' .env | xargs) && \ 
	migrate -path migrations -database "$$DATABASE_URL" version
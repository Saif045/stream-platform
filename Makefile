.PHONY: server mediamtx register login token me channel stream publish streams dev clean-state db-reset

AUTH_USERNAME ?= seif
AUTH_PASSWORD ?= password123
CHANNEL_SLUG ?= channel-$(AUTH_USERNAME)

TABLE ?= users

server:
	go run ./cmd/server

mediamtx:
	mediamtx

register:
	@curl -s -X POST http://localhost:8080/api/auth/register \
	-H "Content-Type: application/json" \
	-d '{"username":"$(AUTH_USERNAME)","password":"$(AUTH_PASSWORD)"}' | jq

login:
	@curl -s -X POST http://localhost:8080/api/auth/login \
	-H "Content-Type: application/json" \
	-d '{"username":"$(AUTH_USERNAME)","password":"$(AUTH_PASSWORD)"}' | jq

token:
	@curl -s -X POST http://localhost:8080/api/auth/login \
	-H "Content-Type: application/json" \
	-d '{"username":"$(AUTH_USERNAME)","password":"$(AUTH_PASSWORD)"}' | jq -r .token

me:
	@TOKEN=$$(make -s token AUTH_USERNAME=$(AUTH_USERNAME) AUTH_PASSWORD=$(AUTH_PASSWORD)); \
	curl -s http://localhost:8080/api/auth/me \
	-H "Authorization: Bearer $$TOKEN" | jq

channel:
	@if [ -f .channel.json ] && [ "$$(jq -r '.id // empty' .channel.json)" != "" ]; then \
		echo "using existing channel:"; \
		cat .channel.json | jq; \
	else \
		TOKEN=$$(make -s token AUTH_USERNAME=$(AUTH_USERNAME) AUTH_PASSWORD=$(AUTH_PASSWORD)); \
		RESPONSE=$$(curl -s -X POST http://localhost:8080/api/channels \
			-H "Content-Type: application/json" \
			-H "Authorization: Bearer $$TOKEN" \
			-d '{"slug":"$(CHANNEL_SLUG)"}'); \
		echo "$$RESPONSE" | jq; \
		ID=$$(echo "$$RESPONSE" | jq -r '.id // empty'); \
		if [ "$$ID" = "" ]; then \
			echo "channel create failed; not writing .channel.json"; \
			exit 1; \
		fi; \
		echo "$$RESPONSE" > .channel.json; \
	fi

stream:
	@if [ ! -f .channel.json ] || [ "$$(jq -r '.id // empty' .channel.json)" = "" ]; then \
		echo "missing valid .channel.json; run make channel first"; \
		exit 1; \
	fi; \
	TOKEN=$$(make -s token AUTH_USERNAME=$(AUTH_USERNAME) AUTH_PASSWORD=$(AUTH_PASSWORD)); \
	CHANNEL_ID=$$(jq -r .id .channel.json); \
	RESPONSE=$$(curl -s -X POST http://localhost:8080/api/live/streams/create \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d "{\"channel_id\":\"$$CHANNEL_ID\"}"); \
	echo "$$RESPONSE" | jq; \
	ID=$$(echo "$$RESPONSE" | jq -r '.id // empty'); \
	if [ "$$ID" = "" ]; then \
		echo "stream create failed; not writing .stream.json"; \
		exit 1; \
	fi; \
	echo "$$RESPONSE" > .stream.json

publish:
	@if [ ! -f .stream.json ] || [ "$$(jq -r '.rtmp_url // empty' .stream.json)" = "" ]; then \
		echo "missing valid .stream.json; run make stream first"; \
		exit 1; \
	fi; \
	ffmpeg -re -stream_loop -1 -i test.mp4 \
	-c copy \
	-f flv \
	"$$(jq -r .rtmp_url .stream.json)"

streams:
	@curl -s http://localhost:8080/api/live/streams | jq

dev: channel stream

clean-state:
	rm -f .channel.json .stream.json

migrate-up:
	. ./.env && migrate -path migrations -database "$$MIGRATION_DATABASE_URL" up

migrate-version:
	. ./.env && migrate -path migrations -database "$$MIGRATION_DATABASE_URL" version

migrate-down:
	. ./.env && migrate -path migrations -database "$$MIGRATION_DATABASE_URL" down 1
	
db-reset-schema:
	sudo -u postgres psql stream_platform -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public AUTHORIZATION stream_migrator;"
# 	$(MAKE) migrate-up
	rm -f .channel.json .stream.json

db-table:
	@. ./.env && psql "$$DATABASE_URL" -c "SELECT * FROM $(TABLE);"

db-table-x:
	@. ./.env && psql "$$DATABASE_URL" -c "\x on" -c "SELECT * FROM $(TABLE);"
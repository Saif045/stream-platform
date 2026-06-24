#!/usr/bin/env bash
set -euo pipefail

BASE="${BASE:-http://localhost:8080}"

TS="$(date +%s)"

USER_A="smoke_user_a_${TS}"
USER_B="smoke_user_b_${TS}"
PASS="password123"
SLUG_A="smoke-${TS}"

echo
echo "== healthz =="
curl -s -i "$BASE/healthz"

echo
echo "== readyz =="
curl -s -i "$BASE/readyz"

echo
echo "== register invalid json -> 400 =="
curl -s -i \
  -X POST "$BASE/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"bad"'

echo
echo "== register user A -> 201 =="
curl -s -i \
  -X POST "$BASE/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER_A\",\"password\":\"$PASS\"}"

echo
echo "== duplicate register user A -> 409 =="
curl -s -i \
  -X POST "$BASE/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER_A\",\"password\":\"$PASS\"}"

echo
echo "== login user A -> 200 =="
TOKEN_A="$(
  curl -s \
    -X POST "$BASE/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USER_A\",\"password\":\"$PASS\"}" \
  | jq -r .token
)"

echo "TOKEN_A length: ${#TOKEN_A}"

echo
echo "== login wrong password -> 401 =="
curl -s -i \
  -X POST "$BASE/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER_A\",\"password\":\"wrong-password\"}"

echo
echo "== me with valid token -> 200 =="
curl -s -i \
  "$BASE/api/auth/me" \
  -H "Authorization: Bearer $TOKEN_A"

echo
echo "== create channel without token -> 401 =="
curl -s -i \
  -X POST "$BASE/api/channels" \
  -H "Content-Type: application/json" \
  -d "{\"slug\":\"$SLUG_A\"}"

echo
echo "== create channel user A -> 201 =="
CHANNEL_JSON="$(
  curl -s \
    -X POST "$BASE/api/channels" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN_A" \
    -d "{\"slug\":\"$SLUG_A\"}"
)"

echo "$CHANNEL_JSON" | jq .
CHANNEL_ID="$(echo "$CHANNEL_JSON" | jq -r .id)"
echo "CHANNEL_ID: $CHANNEL_ID"

echo
echo "== duplicate channel slug -> 409 =="
curl -s -i \
  -X POST "$BASE/api/channels" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_A" \
  -d "{\"slug\":\"$SLUG_A\"}"

echo
echo "== register user B -> 201 =="
curl -s -i \
  -X POST "$BASE/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USER_B\",\"password\":\"$PASS\"}"

echo
echo "== login user B -> 200 =="
TOKEN_B="$(
  curl -s \
    -X POST "$BASE/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USER_B\",\"password\":\"$PASS\"}" \
  | jq -r .token
)"

echo "TOKEN_B length: ${#TOKEN_B}"

echo
echo "== user B creates stream for user A channel -> 403 =="
curl -s -i \
  -X POST "$BASE/api/live/streams/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_B" \
  -d "{\"channel_id\":\"$CHANNEL_ID\"}"

echo
echo "== create stream for missing channel -> 404 =="
curl -s -i \
  -X POST "$BASE/api/live/streams/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_A" \
  -d "{\"channel_id\":\"missing-channel-id\"}"

echo
echo "== user A creates stream for own channel -> 201 =="
STREAM_JSON="$(
  curl -s \
    -X POST "$BASE/api/live/streams/create" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN_A" \
    -d "{\"channel_id\":\"$CHANNEL_ID\"}"
)"

echo "$STREAM_JSON" | jq .
STREAM_ID="$(echo "$STREAM_JSON" | jq -r .id)"
STREAM_KEY="$(echo "$STREAM_JSON" | jq -r .stream_key)"

echo "STREAM_ID: $STREAM_ID"
echo "STREAM_KEY length: ${#STREAM_KEY}"

echo
echo "== user A creates second stream for same channel -> 409 =="
curl -s -i \
  -X POST "$BASE/api/live/streams/create" \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d "{\"channel_id\":\"$CHANNEL_ID\"}"

echo
echo "== list live streams -> 200 and no sensitive fields =="
LIVE_STREAMS_JSON="$(curl -s "$BASE/api/live/streams")"
echo "$LIVE_STREAMS_JSON" | jq .

if echo "$LIVE_STREAMS_JSON" | grep -qE 'stream_key|rtmp_url|output_dir|error'; then
  echo "ERROR: public live streams response leaked sensitive/internal fields"
  exit 1
fi

echo
echo "== list channel streams by slug -> 200 =="
curl -s -i "$BASE/api/channels/slug/$SLUG_A/streams"

echo
echo "== watch page route exists =="
curl -s -i "$BASE/channels/$SLUG_A/watch" | head -n 20

echo
echo "Smoke test complete."
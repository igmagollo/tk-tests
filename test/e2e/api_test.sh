#!/usr/bin/env sh
# Black-box API checks against a running server.
# Usage: BASE_URL=http://localhost:8080 ./test/e2e/api_test.sh
set -eu

BASE_URL="${BASE_URL:-http://localhost:8080}"
failures=0

check() {
	name="$1"
	expected="$2"
	actual="$3"
	if [ "$expected" = "$actual" ]; then
		echo "PASS  $name"
	else
		echo "FAIL  $name: expected [$expected], got [$actual]"
		failures=$((failures + 1))
	fi
}

status() { curl -s -o /dev/null -w '%{http_code}' "$@"; }

echo "Running API checks against $BASE_URL"

check "GET /healthz returns 200" 200 "$(status "$BASE_URL/healthz")"
check "GET /healthz reports ok" '"status":"ok"' \
	"$(curl -s "$BASE_URL/healthz" | grep -o '"status":"ok"' || true)"
check "GET /api/items returns 200" 200 "$(status "$BASE_URL/api/items")"

created=$(curl -s -X POST "$BASE_URL/api/items" -d '{"name":"e2e-widget"}')
check "POST /api/items echoes the name" '"name":"e2e-widget"' \
	"$(echo "$created" | grep -o '"name":"e2e-widget"' || true)"
check "POST /api/items assigns an id" '"id":1' \
	"$(echo "$created" | grep -o '"id":1' || true)"

check "created item appears in the list" '"name":"e2e-widget"' \
	"$(curl -s "$BASE_URL/api/items" | grep -o '"name":"e2e-widget"' || true)"

check "POST /api/items rejects an empty name" 400 \
	"$(status -X POST "$BASE_URL/api/items" -d '{"name":""}')"
check "POST /api/items rejects invalid JSON" 400 \
	"$(status -X POST "$BASE_URL/api/items" -d '{')"

echo
if [ "$failures" -eq 0 ]; then
	echo "All checks passed."
else
	echo "$failures check(s) failed."
	exit 1
fi

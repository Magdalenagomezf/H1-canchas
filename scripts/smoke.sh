#!/usr/bin/env bash
# Smoke test for a deployed backend.
#
# Usage: scripts/smoke.sh <api_base_url> [expected_commit]
#
# Passes when:
#   1. GET /health answers 200 (API up and database reachable), and
#   2. its "commit" matches expected_commit (prefix match, so a short SHA works;
#      a "-dirty" build never matches), and
#   3. GET /spaces answers 200 (a real query through the whole stack).
#
# Retries /health to absorb cold starts and rollouts that still serve the old
# commit. Tune with SMOKE_ATTEMPTS (default 18) and SMOKE_DELAY_SECONDS (default 10).
set -euo pipefail

BASE_URL="${1:?usage: smoke.sh <api_base_url> [expected_commit]}"
BASE_URL="${BASE_URL%/}"
EXPECTED="${2:-}"
ATTEMPTS="${SMOKE_ATTEMPTS:-18}"
DELAY="${SMOKE_DELAY_SECONDS:-10}"

body_file="$(mktemp)"
trap 'rm -f "$body_file"' EXIT

# get <path>: writes the body to $body_file and prints the HTTP status ("000" on network error).
get() {
  curl -s -m 15 -o "$body_file" -w '%{http_code}' "$BASE_URL$1" || true
}

healthy=false
for ((i = 1; i <= ATTEMPTS; i++)); do
  status="$(get /health)"
  commit="$(grep -o '"commit":"[^"]*"' "$body_file" | cut -d'"' -f4 || true)"
  if [[ "$status" == "200" && ( -z "$EXPECTED" || ( "$commit" == "$EXPECTED"* && "$commit" != *-dirty ) ) ]]; then
    healthy=true
    break
  fi
  echo "attempt $i/$ATTEMPTS: /health status=$status commit=${commit:-none} (want ${EXPECTED:-any}); retrying in ${DELAY}s"
  sleep "$DELAY"
done

if [[ "$healthy" != true ]]; then
  echo "FAIL: /health never reported status 200 with the expected commit"
  cat "$body_file"; echo
  exit 1
fi
echo "ok: /health 200, commit $commit"

status="$(get /spaces)"
if [[ "$status" != "200" ]]; then
  echo "FAIL: GET /spaces returned $status"
  cat "$body_file"; echo
  exit 1
fi
echo "ok: /spaces 200"
echo "smoke test passed"

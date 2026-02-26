#!/bin/bash

# Load URLs from file into bloom filter cache via writer_lambda
# Usage: ./load_urls.sh [urls_file] [lambda_host] [batch_size]

URLS_FILE="${1:-urls.txt}"
LAMBDA_HOST="${2:-http://localhost:9000}"
BATCH_SIZE="${3:-100}"

# Lambda RIE invocation endpoint
LAMBDA_ENDPOINT="${LAMBDA_HOST}/2015-03-31/functions/function/invocations"

if [ ! -f "$URLS_FILE" ]; then
    echo "Error: File '$URLS_FILE' not found"
    exit 1
fi

TOTAL_URLS=$(wc -l < "$URLS_FILE" | tr -d ' ')
echo "Loading $TOTAL_URLS URLs from $URLS_FILE via writer_lambda at $LAMBDA_HOST"
echo "Batch size: $BATCH_SIZE"
echo ""

LOADED=0
FAILED=0
BATCH_NUM=0

# Read URLs in batches
while true; do
    # Get next batch of URLs
    BATCH=$(sed -n "$((BATCH_NUM * BATCH_SIZE + 1)),$((( BATCH_NUM + 1) * BATCH_SIZE))p" "$URLS_FILE")

    if [ -z "$BATCH" ]; then
        break
    fi

    # Convert batch to JSON array
    JSON_KEYS=$(echo "$BATCH" | jq -R -s 'split("\n") | map(select(length > 0))')

    # Send batch to writer_lambda (Lambda expects {"keys": [...]})
    RESPONSE=$(curl -s -X POST "${LAMBDA_ENDPOINT}" \
        -H "Content-Type: application/json" \
        -d "{\"keys\": $JSON_KEYS}" \
        -w "\n%{http_code}")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')

    BATCH_COUNT=$(echo "$BATCH" | wc -l | tr -d ' ')

    # Lambda returns {"statusCode": 200, ...} in the body
    STATUS_CODE=$(echo "$BODY" | jq -r '.statusCode // empty' 2>/dev/null)

    if [ "$HTTP_CODE" = "200" ] && [ "$STATUS_CODE" = "200" ]; then
        LOADED=$((LOADED + BATCH_COUNT))
        echo "Batch $((BATCH_NUM + 1)): loaded $BATCH_COUNT URLs (total: $LOADED/$TOTAL_URLS)"
    else
        FAILED=$((FAILED + BATCH_COUNT))
        echo "Batch $((BATCH_NUM + 1)): FAILED - HTTP $HTTP_CODE - $BODY"
    fi

    BATCH_NUM=$((BATCH_NUM + 1))
done

echo ""
echo "=== Summary ==="
echo "Total URLs:  $TOTAL_URLS"
echo "Loaded:      $LOADED"
echo "Failed:      $FAILED"

# Show bloom filter stats (directly from cache service)
CACHE_HOST="${4:-http://localhost:8080}"
echo ""
echo "=== Bloom Filter Stats ==="
curl -s "${CACHE_HOST}/bloom/stats" | jq .

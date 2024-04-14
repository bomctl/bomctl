#!/usr/bin/env bash

set -euo pipefail

readonly APP_ID="874590"

now=$(date +%s)
iat=$((now - 60))  # Issues 60 seconds in the past
exp=$((now + 600)) # Expires 10 minutes in the future

# Install OpenSSL if not installed
if ! command -v openssl &> /dev/null; then
  (apt-get update && apt-get install --yes openssl) &> /dev/null
fi

b64enc() { openssl base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n'; }

header_json='{
  "typ":"JWT",
  "alg":"RS256"
}'

# Header encode
header=$(echo -n "${header_json}" | b64enc)

payload_json='{
  "iat":'"${iat}"',
  "exp":'"${exp}"',
  "iss":'"${APP_ID}"'
}'

# Payload encode
payload=$(echo -n "${payload_json}" | b64enc)

# Signature
header_payload="${header}"."${payload}"
signature=$(
  openssl dgst -sha256 -sign <(echo -n "$GORELEASER_BOT_RSA_PRIVATE_KEY") \
    <(echo -n "${header_payload}") | b64enc
)

# Create JWT
echo "${header_payload}"."${signature}"

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

function b64enc {
  openssl base64 | tr --delete "=" | tr "/+" "_-" | tr --delete "\n"
}

# Encode header
header=$(printf '{"typ": "JWT", "alg": "RS256"}' | b64enc)

# Encode payload
payload=$(printf '{"iat": "%s", "exp": "%s", "iss": "%s"}' $iat $exp $APP_ID | b64enc)

# Signature
header_payload="${header}.${payload}"
signature=$(
  printf %s "${header_payload}" |
    openssl dgst -sha256 -sign <(printf %s "${GORELEASER_BOT_RSA_PRIVATE_KEY}") |
    b64enc
)

# Create JWT
echo "${header_payload}.${signature}"

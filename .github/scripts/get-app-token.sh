#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]:-$0}")" &> /dev/null && pwd)"
readonly SCRIPT_DIR

# Install jq to parse JSON responses
if ! command -v jq &> /dev/null; then
  (apt-get update && apt-get install --yes jq) &> /dev/null
fi

jwt=$("$SCRIPT_DIR/generate-jwt.sh")

curl_args=(
  --fail --silent --location
  --header "Accept: application/vnd.github+json"
  --header "X-GitHub-Api-Version: 2022-11-28"
)

# Get ID of GitHub App installation
installation_id=$(curl "${curl_args[@]}" \
  --header "Authorization: Bearer $jwt" \
  --url https://api.github.com/repos/bomctl/bomctl/installation | jq --raw-output .id)

# Get installation access token
token=$(curl --request POST "${curl_args[@]}" \
  --header "Authorization: Bearer $jwt" \
  --url "https://api.github.com/app/installations/$installation_id/access_tokens" | jq --raw-output .token)

echo "$token"

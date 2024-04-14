#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]:-$0}")" &> /dev/null && pwd)"
readonly SCRIPT_DIR

# Install jq to parse JSON responses
if ! command -v jq &> /dev/null; then
  (apt-get update && apt-get install --yes jq) &> /dev/null
fi

jwt=$("$SCRIPT_DIR/generate-jwt.sh")

# Get ID of GitHub App installation
installation_id="$(curl --fail --silent --location --request GET \
  --header "Accept: application/vnd.github+json" \
  --header "Authorization: Bearer $jwt" \
  --header "X-GitHub-Api-Version: 2022-11-28" \
  --url https://api.github.com/repos/bomctl/bomctl/installation | jq --raw-output .id)"

# Get access token
token="$(curl --fail --silent --location --request POST \
  --header "Accept: application/vnd.github+json" \
  --header "Authorization: Bearer $jwt" \
  --header "X-GitHub-Api-Version: 2022-11-28" \
  --url "https://api.github.com/app/installations/$installation_id/access_tokens" | jq --raw-output .token)"

# Create a tag as reference
curl --fail --silent --location --request POST \
  --header "Accept: application/vnd.github+json" \
  --header "Authorization: Bearer $token" \
  --header "X-GitHub-Api-Version: 2022-11-28" \
  --url https://api.github.com/repos/bomctl/bomctl/git/refs \
  --data '{
    "ref": "refs/tags/'"$NEXT_VERSION"'",
    "sha": "'"$GITHUB_REF"'"
  }'

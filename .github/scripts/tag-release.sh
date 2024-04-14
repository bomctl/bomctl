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
installation_id="$(curl --request GET "${curl_args[@]}" \
  --header "Authorization: Bearer $jwt" \
  --url https://api.github.com/repos/bomctl/bomctl/installation | jq --raw-output .id)"

# Get access token
token="$(curl --request POST "${curl_args[@]}" \
  --header "Authorization: Bearer $jwt" \
  --url "https://api.github.com/app/installations/$installation_id/access_tokens" | jq --raw-output .token)"

# Create a tag
curl --request POST --output /dev/null "${curl_args[@]}" \
  --header "Authorization: Bearer $token" \
  --url https://api.github.com/repos/bomctl/bomctl/git/tags \
  --data '{
      "tag": "'"$NEXT_VERSION"'",
      "message": "'"$NEXT_VERSION"'",
      "object": "'"$GITHUB_SHA"'",
      "type": "commit",
      "tagger": {
        "name": "bomctl-goreleaser-bot[bot]",
        "email": "166692013+bomctl-goreleaser-bot[bot]@users.noreply.github.com"
      }
    }'

# Create a reference to the tag
curl --request POST --output /dev/null "${curl_args[@]}" \
  --header "Authorization: Bearer $token" \
  --url https://api.github.com/repos/bomctl/bomctl/git/refs \
  --data '{
    "ref": "refs/tags/'"$NEXT_VERSION"'",
    "sha": "'"$GITHUB_SHA"'"
  }'

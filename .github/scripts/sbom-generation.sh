#!/usr/bin/env bash

set -euo pipefail

# ANSI color escape codes
readonly YELLOW="\033[33m"
readonly RESET="\033[0m"

function usage {
  echo -e "${YELLOW}Usage: sbom-generation.sh { artifact } { cyclonedx|spdx } { document }${RESET}"
  exit 1
}

if [[ $# -ne 3 ]]; then usage; fi

syft scan "${1}" --output "${2}-json=syft-${3}"

# Use jq to merge additional metadata into the sboms
jq --slurp '.[0] * .[1]' "syft-${3}" "../.github/sbom_metadata/metadata.${2}.json" > "${3}"

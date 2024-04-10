#!/usr/bin/env bash

set -euo pipefail

version="$(git describe --tags --abbrev=0)"
version="${version#v}"

# Split version string into components
IFS="-" read -r version dev <<< "$version"
IFS="." read -r major minor patch <<< "$version"

sed -E --in-place \
  --expression='s/^(\tVersionMajor = )[0-9]+$/\1'"$major"'/' \
  --expression='s/^(\tVersionMinor = )[0-9]+$/\1'"$minor"'/' \
  --expression='s/^(\tVersionPatch = )[0-9]+$/\1'"$patch"'/' \
  --expression='s/^(\tVersionDev = )".*"$/\1"-'"$dev"'"/' \
  cmd/version.go

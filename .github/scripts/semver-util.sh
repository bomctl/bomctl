#!/usr/bin/env bash

set -euo pipefail

# ANSI color escape codes
readonly BOLD="\033[1m"
readonly GREEN="\033[32m"
readonly YELLOW="\033[33m"
readonly CYAN="\033[36m"
readonly RED="\033[31m"
readonly RESET="\033[0m"

function usage {
  echo -e "${YELLOW}Usage: semver-util.sh { next | print }${RESET}"
  exit 1
}

function get_next {
  local current_branch
  current_branch="$(git branch --show-current)"

  if [[ $current_branch =~ ^alpha|dev$ ]]; then
    svu prerelease --pre-release "$current_branch"
  else
    svu next
  fi
}

function print_version {
  local next_version major minor patch pre

  next_version="$(get_next)"

  # Split next_version string into components
  IFS="-" read -r version pre <<< "${next_version#v}"
  IFS="." read -r major minor patch <<< "$version"

  release_info=(
    ""
    "-------+---------"
    " ${BOLD}${CYAN}type${RESET}  | ${BOLD}${CYAN}value${RESET}"
    "-------+---------"
    " ${BOLD}major${RESET} | ${GREEN}${major}${RESET}"
    " ${BOLD}minor${RESET} | ${GREEN}${minor}${RESET}"
    " ${BOLD}patch${RESET} | ${GREEN}${patch}${RESET}"
    " ${BOLD}pre${RESET}   | ${YELLOW}${pre}${RESET}"
    "-------+---------"
    ""
  )

  printf "%b\n" "${release_info[@]}"
}

if [[ $# -eq 0 ]]; then
  usage
fi

case $1 in
  next)  get_next      ;;
  print) print_version ;;
  *)     echo -e "${RED}Unknown argument: ${1}${RESET}" && usage ;;
esac

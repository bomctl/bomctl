# -------------------------------------------------------
# SPDX-FileCopyrightText: Copyright © 2024 bomctl a Series of LF Projects, LLC
# SPDX-FileName: Makefile
# SPDX-FileType: SOURCE
# SPDX-License-Identifier: Apache-2.0
# -------------------------------------------------------
MAKEFILE ?= ${abspath ${firstword ${MAKEFILE_LIST}}}

# ANSI color escape codes
BOLD   := \033[1m
CYAN   := \033[38;5;51m
GREEN  := \033[38;5;46m
ORANGE := \033[38;5;214m
YELLOW := \033[38;5;226m
RED    := \033[38;5;196m
RESET  := \033[0m

# Default system architecture
ARCH ?= amd64
OS ?= linux

VERSION := ${shell git describe --tags --abbrev=0}
VERSION ?= undefined

GIT_SHA := ${shell git rev-parse HEAD}
GIT_SHA ?= undefined

BUILD_DATE := ${shell date -u +'%Y-%m-%dT%H:%M:%SZ'}

VERSION_PARTS := ${subst ., ,${firstword ${subst -, ,${VERSION:v%=%}}}}

VERSION_MAJOR := ${word 1,${VERSION_PARTS}}
VERSION_MINOR := ${word 2,${VERSION_PARTS}}
VERSION_PATCH := ${word 3,${VERSION_PARTS}}

VERSION_PRE := ${word 2,${subst -, ,${VERSION}}}
VERSION_PRE := ${if ${VERSION_PRE},-${VERSION_PRE},}

LDFLAGS := -s -w \
  -X=github.com/bomctl/bomctl/cmd.VersionMajor=${VERSION_MAJOR} \
  -X=github.com/bomctl/bomctl/cmd.VersionMinor=${VERSION_MINOR} \
  -X=github.com/bomctl/bomctl/cmd.VersionPatch=${VERSION_PATCH} \
  -X=github.com/bomctl/bomctl/cmd.VersionPre=${VERSION_PRE} \
  -X=github.com/bomctl/bomctl/cmd.BuildDate=${BUILD_DATE}

GOPATH ?= ${shell go env GOPATH}
GOLANGCI_LINT_INSTALL := https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh
GOLANGCI_LINT_VERSION := v1.59.1

SHFMT_VERSION := v3.8.0

SHELLCHECK_VERSION  := v0.10.0
SHELLCHECK_FILENAME := shellcheck-${SHELLCHECK_VERSION}

ifeq (${OS},Windows_NT)
	OS := windows
	SHELLCHECK_FILENAME := ${addsuffix .zip,${SHELLCHECK_FILENAME}}

	ifeq (${PROCESSOR_ARCHITECTURE},x86)
		ARCH := i386
	endif
else
	uname_s := ${shell uname -s}
	uname_p := ${shell uname -p}

	ifeq (${uname_s},Darwin)
		OS := macos
		SHELLCHECK_FILENAME := ${addsuffix .darwin,${SHELLCHECK_FILENAME}}
	else
		SHELLCHECK_FILENAME := ${addsuffix .linux,${SHELLCHECK_FILENAME}}
	endif

	ifeq (${uname_p},arm)
		ARCH := arm64
		SHELLCHECK_FILENAME := ${addsuffix .aarch64,${SHELLCHECK_FILENAME}}
	else
		SHELLCHECK_FILENAME := ${addsuffix .x86_64,${SHELLCHECK_FILENAME}}
	endif

	SHELLCHECK_FILENAME := ${addsuffix .tar.xz,${SHELLCHECK_FILENAME}}
endif

TARGET_BIN := ${PWD}/build/bomctl-${OS}-${ARCH}

ifeq (${OS},windows)
	TARGET_BIN := ${addsuffix .exe,${TARGET_BIN}}
endif

.PHONY: all

#@ Tools
help: # Display this help
	@awk 'BEGIN {FS = ":.*#"; printf "\n${YELLOW}Usage: make <target>${RESET}\n"} \
		/^[a-zA-Z_0-9-]+:.*?#/ { printf "  ${CYAN}%-20s${RESET} %s\n", $$1, $$2 } \
		/^#@/ { printf "\n${BOLD}%s${RESET}\n", substr($$0, 4) }' ${MAKEFILE_LIST} && echo

.PHONY: clean
clean: # Clean the working directory
	@${RM} -r dist
	@find ${PWD} -name "*.log" -exec ${RM} {} \;

SHELLCHECK_DOWNLOAD_URL := https://github.com/koalaman/shellcheck/releases/download/${SHELLCHECK_VERSION}/${SHELLCHECK_FILENAME}

.PHONY: install
install: # Install dev tools
	@mkdir -p .bin

	@if ! command -v golangci-lint > /dev/null; then \
	  printf "${YELLOW}golangci-lint not found. Installing... ${RESET}\n\n"; \
	  curl --fail --silent --show-error --location ${GOLANGCI_LINT_INSTALL} | \
	    sh -s -- -b ${GOPATH}/bin ${GOLANGCI_LINT_VERSION}; \
	fi

	@if [ ! -f .bin/shellcheck ]; then \
	  printf "${YELLOW}shellcheck not found. Installing... ${RESET}\n\n"; \
	  if [ ${OS} = linux ] || [ ${OS} = macos ]; then \
	  	curl --fail --silent --show-error --location --url ${SHELLCHECK_DOWNLOAD_URL} | \
		  tar --extract --xz --directory .bin --strip-components=1 shellcheck-${SHELLCHECK_VERSION}/shellcheck ; \
	  elif [ {OS}$ = windows ]; then \
	  	curl --fail --silent --show-error --location --url ${SHELLCHECK_DOWNLOAD_URL} --output temp.zip; \
	    unzip temp.zip -d .bin; \
		rm temp.zip; \
	  fi; \
	fi

	@if ! command -v shfmt > /dev/null; then \
	  printf "${YELLOW}shfmt not found. Installing... ${RESET}\n\n"; \
	  go install mvdan.cc/sh/v3/cmd/shfmt@${SHFMT_VERSION}; \
	fi

	@printf "${YELLOW}Development tools installed${RESET}\n\n"

#@ Build
define gobuild
	CGO_ENABLED=0 GOOS=${1} GOARCH=${2} go build -trimpath -o dist/bomctl-${1}-${2}${3} -ldflags="${LDFLAGS}"
endef

.PHONY: build-linux-amd
build-linux-amd: # Build for Linux on AMD64
	${call gobuild,linux,amd64}

.PHONY: build-linux-arm
build-linux-arm: # Build for Linux on ARM
	${call gobuild,linux,arm64}

.PHONY: build-macos-intel
build-macos-intel: # Build for macOS on AMD64
	${call gobuild,darwin,amd64}

.PHONY: build-macos-apple
build-macos-apple: # Build for macOS on ARM
	${call gobuild,darwin,arm64}

.PHONY: build-windows-amd
build-windows-amd: # Build for Windows on AMD64
	${call gobuild,windows,amd64,.exe}

.PHONY: build-windows-arm
build-windows-arm: # Build for Windows on ARM
	${call gobuild,windows,arm64,.exe}

.PHONY: build-linux
build-linux: build-linux-amd build-linux-arm # Build for Linux on AMD64 and ARM

.PHONY: build
build: build-linux-amd build-linux-arm build-macos-intel build-macos-apple build-windows-amd build-windows-arm ## Build the CLI

#@ Lint
define run-lint
	@export PATH="$${PATH}:$${PWD}/.bin"; \
	if command -v ${1} > /dev/null; then \
	  printf "Running ${CYAN}${1} ${2}${RESET}\n\n"; \
	  ${1} ${2}; \
	else \
	  printf "${YELLOW}${1} not found, please install and run the command again.${RESET}\n"; \
	fi
endef

.PHONY: lint-go
lint-go: # Lint Golang code files
	${call run-lint,golangci-lint,run --verbose}
	${call run-lint,.github/scripts/check-go-headers.sh}

.PHONY: lint-go-fix
lint-go-fix: # Fix golangci-lint findings
	${call run-lint,golangci-lint,run --fix --verbose}

.PHONY: lint-markdown
lint-markdown: # Lint markdown files
	${call run-lint,markdownlint-cli2,${shell git ls-files '*.md'}}

.PHONY: lint-markdown-fix
lint-markdown-fix: # Fix markdown lint findings
	${call run-lint,markdownlint-cli2,${shell git ls-files '*.md'} --fix}

.PHONY: lint-shell
lint-shell: # Lint shell scripts
	${call run-lint,shellcheck,${shell git ls-files '*.sh'}}
	${call run-lint,shfmt,--diff --simplify ${shell git ls-files '*.sh'}}

.PHONY: lint-yaml
lint-yaml: # Lint YAML files
	${call run-lint,yamllint,${shell git ls-files '*.yml' '*.yaml'}}

.PHONY: lint
lint: lint-go lint-markdown lint-shell lint-yaml # Lint Golang code, markdown, shell script, and YAML files

.PHONY: lint-fix
lint-fix: lint-go-fix lint-markdown-fix lint-shell lint-yaml # Lint Golang code, markdown, shell script, and YAML files, apply fixes where possible

#@ Test
define coverage-report
	@printf "${CYAN}"; \
	echo "+----------------------------------------------------------------------------------------+"; \
	echo "|    COVERAGE REPORT                                                                     |"; \
	echo "+----------------------------------------------------------------------------------------+"; \
	printf "${RESET}\n"

	@go tool cover -func=coverage.out | \
	  awk -- '{ \
	    sub("github.com/bomctl/bomctl/", "", $$1); \
	    percent = +$$3; sub("%", "", percent); \
	    if (percent < 50.00) color = "${RED}"; \
	    else if (percent < 80.00) color = "${ORANGE}"; \
	    else if (percent < 100.00) color = "${YELLOW}"; \
	    else color = "${GREEN}"; \
	    fmtstr = $$1 == "total:" ? "\n%s%s\t%s\t%s%s\n" : "%s%-48s %-32s %.1f%%%s\n"; \
	    printf fmtstr, color, $$1, $$2, $$3, "${RESET}" \
	  }'
endef

.PHONY: test-unit
test-unit: # Run unit tests
	@printf "Running unit tests for ${BOLD}${CYAN}bomctl${RESET}..."
	@go test -failfast -v -coverprofile=coverage.out -covermode=atomic -short ./...
	@printf "${BOLD}${GREEN}DONE${RESET}\n\n"

	${call coverage-report}

.PHONY: test-e2e
test-e2e: # Run unit tests
	@printf "Running end-to-end tests for ${BOLD}${CYAN}bomctl${RESET}..."
	@go test -failfast -v ./internal/e2e/...
	@printf "${BOLD}${GREEN}DONE${RESET}\n\n"

	${call coverage-report}

.PHONY: test
test: # Run all tests
	@printf "Running all tests for ${BOLD}${CYAN}bomctl${RESET}..."
	@go test -failfast -v -coverprofile=coverage.out -covermode=atomic ./...
	@printf "${BOLD}${GREEN}DONE${RESET}\n\n"

	${call coverage-report}

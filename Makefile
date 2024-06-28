# -------------------------------------------------------
# SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
# SPDX-FileName: Makefile
# SPDX-FileType: SOURCE
# SPDX-License-Identifier: Apache-2.0
# -------------------------------------------------------
MAKEFILE ?= ${abspath ${firstword ${MAKEFILE_LIST}}}

# ANSI color escape codes
BOLD   := \033[1m
CYAN   := \033[36m
YELLOW := \033[33m
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

#@ Test
.PHONY: test-unit
test-unit: # Run unit tests
	go test -failfast -v -coverprofile=coverage.out -covermode=atomic ./...

.PHONY: test
test: test-unit # Run all tests

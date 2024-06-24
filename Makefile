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

ifeq (${OS},Windows_NT)
	OS := windows

	ifeq (${PROCESSOR_ARCHITECTURE},x86)
		ARCH := i386
	endif
else
	uname_s := ${shell uname -s}
	uname_p := ${shell uname -p}

	ifeq (${uname_s},Darwin)
		OS := macos
	endif

	ifeq (${uname_p},arm)
		ARCH := arm64
	endif
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
	@if command -v ${1} &> /dev/null; then \
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
	${call run-lint,markdownlint-cli2,'**/*.md'}

.PHONY: lint-markdown-fix
lint-markdown-fix: # Fix markdown lint findings
	${call run-lint,markdownlint-cli2,'**/*.md' --fix}

.PHONY: lint-yaml
lint-yaml: # Lint YAML files
	${call run-lint,yamllint,.}

.PHONY: lint
lint: lint-go lint-markdown lint-yaml # Lint Golang code, markdown, and YAML files

#@ Test
.PHONY: test-unit
test-unit: # Run unit tests
	go test -failfast -v -coverprofile=coverage.out -covermode=atomic ./...

.PHONY: test
test: test-unit # Run all tests

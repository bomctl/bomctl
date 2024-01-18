# -------------------------------------------------------
# SPDX-FileCopyrightText: Copyright © 2024 bomctl authors
# SPDX-FileName: Makefile
# SPDX-FileType: SOURCE
# SPDX-License-Identifier: Apache-2.0
# -------------------------------------------------------
BASH := ${shell type -p bash}
SHELL := ${BASH}
MAKEFILE ?= ${abspath ${firstword ${MAKEFILE_LIST}}}

# Default system architecture
ARCH ?= amd64
OS ?= linux

# ANSI color escape codes
BOLD ?= \033[1m
CYAN ?= \033[36m
GREEN ?= \033[32m
RED ?= \033[31m
YELLOW ?= \033[33m
NC ?= \033[0m # No Color

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

CLI_VERSION ?= $(if $(shell git describe --tags),$(shell git describe --tags),"UnknownVersion")
GIT_SHA := $(if $(shell git rev-parse HEAD),$(shell git rev-parse HEAD),"")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

.PHONY: all build clean help format test
.SILENT: clean

#@ Tools
help: # Display this help
	@awk 'BEGIN {FS = ":.*#"; printf "\n${YELLOW}Usage: make <target>${NC}\n"} \
		/^[a-zA-Z_0-9-]+:.*?#/ { printf "  ${CYAN}%-20s${NC} %s\n", $$1, $$2 } \
		/^#@/ { printf "\n${BOLD}%s${NC}\n", substr($$0, 4) }' ${MAKEFILE} && echo

clean: # Clean the working directory
	${RM} -r build
	find ${PWD} -name "*.log" -exec ${RM} {} \;

lint: # Lint Golang code files
	golangci-lint run --verbose

lint-fix: # Fix linter findings
	golangci-lint run --fix --verbose

#@ Build
define gobuild
	CGO_ENABLED=0 GOOS=${1} GOARCH=${2} go build -trimpath -o build/bomctl-${1}-${2}${3}
endef

build-linux-amd: # Build for Linux on AMD64
	${call gobuild,linux,amd64}

build-linux-arm: # Build for Linux on ARM
	${call gobuild,linux,arm64}

build-macos-intel: # Build for macOS on AMD64
	${call gobuild,darwin,amd64}

build-macos-apple: # Build for macOS on ARM
	${call gobuild,darwin,arm64}

build-windows-amd: # Build for Windows on AMD64
	${call gobuild,windows,amd64,.exe}

build-windows-arm: # Build for Windows on ARM
	${call gobuild,windows,arm64,.exe}

build-linux: build-linux-amd build-linux-arm # Build for Linux on AMD64 and ARM

build: build-linux-amd build-linux-arm build-macos-intel build-macos-apple build-windows-amd build-windows-arm ## Build the CLI

# Development

This document describes the process for building and running `bomctl`.

## Requirements

- [go command](https://go.dev/dl/)
- [make](https://www.gnu.org/software/make/manual/make.html)
- [golangci-lint](https://golangci-lint.run/welcome/install/#local-installation)

## Getting Started

Clone `bomctl` repository

``` shell
git clone https://github.com/bomctl/bomctl.git
cd bomctl
```

Build using the `Makefile`

``` shell
make build
```

Lint using `golangci-lint`

``` shell
make lint

... and/or ...

make lint-fix
```

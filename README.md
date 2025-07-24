# go-project-boilerplate

[![go](https://github.com/sters/go-project-boilerplate/workflows/Go/badge.svg)](https://github.com/sters/go-project-boilerplate/actions?query=workflow%3AGo)
[![coverage](docs/coverage.svg)](https://github.com/sters/go-project-boilerplate)
[![go-report](https://goreportcard.com/badge/github.com/sters/go-project-boilerplate)](https://goreportcard.com/report/github.com/sters/go-project-boilerplate)

My go project boilerplate.

## Includes

- Makefile
  - run
  - test
  - cover
  - Tools install from `./tools/tools.go`
- Github Actions
  - Go
    - Lint by golangcilint
    - Run test and generate coverage report with octocov
  - Release
    - Make release when vX.X.X tag is added by goreleaser.
- README
  - Badge: Github Actions/Go
  - Badge: Octocov Coverage
  - Badge: Go Report

## TODO when use this

- [ ] Change run task in `Makefile` if needed
- [ ] Change package name in `go.mod`
- [ ] Update `README.md`

---

## Install

```shell
go install github.com/sters/go-project-boilerplate@latest
```

or use specific version from [Releases](https://github.com/sters/go-project-boilerplate/releases).

## Usage

....


export GOBIN := $(PWD)/bin
export PATH := $(GOBIN):$(PATH)

.PHONY: run
run:
	go run main.go $(ARGS)

.PHONY: lint
lint:
	go tool golangci-lint run -v ./...

.PHONY: lint-fix
lint-fix:
	go tool golangci-lint run --fix -v ./...

.PHONY: test
test:
	go test -v -race ./...

.PHONY: cover
cover:
	go test -v -race -coverpkg=./... -coverprofile=coverage.out ./...

.PHONY: tidy
tidy:
	go mod tidy

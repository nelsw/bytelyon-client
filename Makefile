.PHONY: fmt lint test deps clean

KEY=
LOG=debug
SRC=$(shell find . -name "*.go")
URL=http://localhost:80

#
# App Commands
#
# all: periodically polls and works in sequence all workable bots
# one: query the first workable bot, work it, and exit
# sub: subscribe to the server for work when it's ready
#
all:
	@make ƒø name=all && go run ./cmd/all/main.go || true
one:
	@make ƒø name=one && go run ./cmd/one/main.go || true
sub:
	@make ƒø name=sub && go run ./cmd/sub/main.go || true

clean:
	@make ƒø name=clean
	@rm -f .storage ./bin/app
	@make ƒç name=clean

build:
	@make ƒø name=build
	@go generate ./...
	@rm -f ./bin/app
	@go build -o ./bin/app ./cmd/app
	@make ƒç name=build

fmt:
	@make ƒø name=fmt
	@test -z $(shell gofmt -l $(SRC)) || (gofmt -d $(SRC); exit 1)
	@make ƒç name=fmt

lint:
	@make ƒø name=lint
	@golangci-lint run -v
	@make ƒç name=lint

test: deps
	@make ƒø name=test
	@godotenv -f .env go test -v ./...
	@make ƒç name=test

rich: deps
	@make ƒø name=rich
	@godotenv -f .env richgo test -v ./...
	@make ƒç name=rich

deps: install
	@make ƒø name=deps
	@go mod tidy
	@go get -v ./...
	@go mod tidy
	@make ƒç name=deps

install:
	@go get -u github.com/kyoh86/richgo
	@go install github.com/joho/godotenv/cmd/godotenv@latest

ƒø:
	@printf "\033[1;94m❯\033[0m %s [\033[1;94m%s\033[0m]\n" "∙∙∙" "${name}"
ƒç:
	@printf "\033[1;92m❯\033[0m %s [\033[1;92m%s\033[0m]\n" "∙∙∙" "${name}"
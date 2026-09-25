.PHONY: fmt lint test deps clean

KEY=
LOG=debug
SRC=$(shell find . -name "*.go")
URL=http://localhost:80

ƒø:
	@printf "\033[1;94m❯\033[0m %s [\033[1;94m%s\033[0m]\n" "∙∙∙" "${name}"
ƒç:
	@printf "\033[1;92m❯\033[0m %s [\033[1;92m%s\033[0m]\n" "∙∙∙" "${name}"

clean:
	@make ƒø name=clean
	@rm -f .storage ./bin/app
	@make ƒç name=clean

app:
	@make ƒø name=app
	@go run ./cmd/app/main.go -log=$(LOG) -url=$(URL) -key=$(KEY)
	@make ƒç name=app

build:
	@make ƒø name=build
	@go generate ./...
	@rm -f ./bin/app
	@go build -o ./bin/app ./cmd/app
	@make ƒç name=build

it: build
	@make ƒø name=it
	@./cmd/app -log=$(LOG) -url=$(URL) -key=$(KEY)
	@make ƒç name=it

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

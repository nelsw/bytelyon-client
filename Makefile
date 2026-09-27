.PHONY: fmt lint test deps clean

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@$(MAKE) banner
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: project-path
project-path: ## Print the project path.
	@echo $(PROJECT_PATH)

##@ Run
tui: ## launch the interactive TUI for running and inspecting make targets
	@go run ./cmd/tui
all: ## periodically polls and works in sequence all workable bots
	@$(MAKE) it APP=all || true
one: ## query the first workable bot, work it, and exit
	@$(MAKE) it APP=one || true
sub: ## subscribe to the server for work when it's ready
	@$(MAKE) it APP=sub || true
it: banner ## helper target for aforementioned targets, requires argument 'APP=<all|one|sub>'
	@$(MAKE) ƒø name=run-$(APP)
	@go run ./cmd/$(APP)/main.go
	@$(MAKE) ƒç name=run-$(APP)

##@ Project
lint: ## checks formats all go files
	@$(MAKE) ƒø name=lint
	@golangci-lint run -v
	@$(MAKE) ƒç name=lint

##@ Source
clean: ## removes all files from .storage directories and clears the .bin folder
	@$(MAKE) ƒø name=clean
	@rm -rf ./bin/all ./bin/one ./bin/sub .storage/*/*
	@$(MAKE) ƒç name=clean
install: clean ## gets and installs all deps for test/dev/prod before a tidy
	@$(MAKE) ƒø name=deps
	@go get -u github.com/kyoh86/richgo
	@go install github.com/joho/godotenv/cmd/godotenv@latest
	@go get -v ./...
	@go mod tidy
	@$(MAKE) ƒç name=deps
build: install ## generate code (jic) and build the executable
	@$(MAKE) ƒø name=build-$(APP)
	@go generate ./...
	@go build -o ./bin/$(APP) ./cmd/$(APP)
	@$(MAKE) ƒç name=build-$(APP)

##@ Test
test: banner install ## verbose richgo test; requires an .env file
	@$(MAKE) ƒø name=rich
	@godotenv -f .env richgo test -coverprofile=coverage.out -v ./...
	@gocovsh
	@$(MAKE) ƒç name=rich

ƒø:
	@printf "\n\033[1;95m❯\033[0m %s [\033[1;95m%s\033[0m]\n" "∙∙∙" "${name}"
ƒç:
	@printf "\033[1;92m❯\033[0m %s [\033[1;92m%s\033[0m]" "∙∙∙" "${name}"
banner:
	@printf "\n[1;93m"
	@printf "\n* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * "
	@printf "\n*                                                                           * "
	@printf "\n*[1;94m    ██████╗ ██╗   ██╗████████╗███████╗██╗  ██╗   ██╗ ██████╗ ███╗   ██╗    \033[1;93m* "
	@printf "\n*[1;94m    ██╔══██╗╚██╗ ██╔╝╚══██╔══╝██╔════╝██║  ╚██╗ ██╔╝██╔═══██╗████╗  ██║    \033[1;93m* "
	@printf "\n*[1;94m    ██████╔╝ ╚████╔╝    ██║   █████╗  ██║   ╚████╔╝ ██║   ██║██╔██╗ ██║    \033[1;93m* "
	@printf "\n*[1;94m    ██╔══██╗  ╚██╔╝     ██║   ██╔══╝  ██║    ╚██╔╝  ██║   ██║██║╚██╗██║    \033[1;93m* "
	@printf "\n*[1;94m    ██████╔╝   ██║      ██║   ███████╗███████╗██║   ╚██████╔╝██║ ╚████║    \033[1;93m* "
	@printf "\n*[1;94m    ╚═════╝    ╚═╝      ╚═╝   ╚══════╝╚══════╝╚═╝    ╚═════╝ ╚═╝  ╚═══╝    \033[1;93m* "
	@printf "\n*                                                                           * "
	@printf "\n* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * "
	@printf "\n[0m"

include .env

BIN="./bin"
SRC=$(shell find . -name "*.go")


ifeq (, $(shell which golangci-lint))
$(warning "could not find golangci-lint in $(PATH), run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.7.2")
endif

.PHONY: fmt lint test install_deps clean


build:
	@printf "➜  %s  %s [\033[35m%s\033[0m]" "🛠" "build" ${name}
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap ./cmd/${name}/main.go
	@zip -r9 main.zip bootstrap > /dev/null
	@printf "  ✅\n"

clean:
	@printf "➜  %s  %s [\033[35m%s\033[0m]" "🧽" "clean" "*"
	@rm -f ./bootstrap
	@rm -f ./main.zip
	@rm -f ./cp.out
	@printf "  ✅\n"

create: build
	@printf "➜  %s  %s [\033[35m%s\033[0m]" "💽" "create" ${name}
	@aws lambda create-function \
		--function-name bytelyon-${name} \
		--runtime "provided.al2023" \
		--role ${AWS_IAM_ROLE} \
		--architectures arm64 \
		--handler "bootstrap" \
		--zip-file "fileb://./main.zip" \
		--memory-size "512" \
		--timeout "30" \
		--publish \
		--environment "Variables={$(shell tr '\n' ',' < .env)}" > /dev/null
	@printf "  ✅\n"
	@make clean

delete:
	@printf "➜  %s  %s [\033[35m%s\033[0m]" "🗑️" "delete" ${name}
	@aws lambda delete-function --function-name bytelyon-${name} | jq
	@printf "  ✅\n"

list:
	@printf "➜  %s  %s [\033[35m./aws/ꟛ/%s\033[0m]" "📋" "list"
	@aws lambda list-functions --no-paginate \
	| jq '.Functions.[] | {name: .FunctionName, updated: .LastModified, environment: .Environment.Variables}'

logs:
	@printf "➜  %s  %s [\033[35m%s\033[0m]" "👀" "logs" ${name}
	open "https://us-east-1.console.aws.amazon.com/cloudwatch/home#logStream:group=/aws/lambda/bytelyon-${name}"

publish:
	@printf "➜  %s  %s [\033[35m%s\033[0m]" "🌐" "publish" ${name}
	@aws lambda create-function-url-config --function-name bytelyon-${name} --auth-type NONE > /dev/null
	@aws lambda add-permission \
    		--function-name bytelyon-${name} \
    		--action lambda:InvokeFunctionUrl \
    		--principal "*" \
    		--statement-id FunctionURLAllowPublicAccess \
    		--function-url-auth-type NONE > /dev/null
	@printf "  ✅\n"
	@make url

#test: clean
#	@printf "➜  %s  %s [\033[35m%s\033[0m]\n\n" "📊" "test" "*"
#	@go test -covermode=atomic -coverpkg=./... -coverprofile=cp.out ./...  > /dev/null
#	@sed -i '' -e '/bytelyon-functions\/cmd\//d' cp.out
#	@sed -i '' -e '/bytelyon-functions\/test\//d' cp.out
#	@go tool cover -func=cp.out
#	@go tool cover -html=cp.out

unpublish:
	@printf "➜  %s  %s [\033[35m%s\033[0m]" "⛔️" "unpublish" ${name}
	@aws lambda remove-permission --function-name bytelyon-${name} --statement-id FunctionURLAllowPublicAccess
	@aws lambda delete-function-url-config --function-name bytelyon-${name}
	@printf "  ✅\n"

update: build
	@printf "➜  %s  %s [\033[35m%s\033[0m]" "💾" "update" ${name}
	@aws lambda update-function-configuration \
    		--function-name bytelyon-${name} \
    		--role ${AWS_IAM_ROLE} \
    		--environment "Variables={$(shell tr '\n' ',' < .env)}" > /dev/null
	@aws lambda update-function-code --zip-file fileb://./main.zip --function-name bytelyon-${name} > /dev/null
	@printf "  ✅\n"
	@make clean

url:
	@printf "➜  %s  %s [\033[35m%s\033[0m]\n" "🛜" "url" ${name}
	@aws lambda get-function-url-config --function-name bytelyon-${name} | jq '.FunctionUrl'

invoke:
	@printf "➜  %s  %s [\033[35m%s\033[0m]\n" "🐻" "invoke" ${name}
	@aws lambda invoke \
		--function-name bytelyon-browser \
		--cli-binary-format raw-in-base64-out \
		--payload '{ "url": "https://google.com/search?q=corsair+marine+970" }' \
		response.json > /dev/null
	@printf "  ✅\n"





#default: all
#
#all: fmt test

fmt:
	$(info ******************** checking formatting ********************)
	@test -z $(shell gofmt -l $(SRC)) || (gofmt -d $(SRC); exit 1)

lint:
	$(info ******************** running lint tools ********************)
	golangci-lint run -v

test: install_deps
	$(info ******************** running tests ********************)
	go test -v ./...

richtest: install_deps
	$(info ******************** running tests with kyoh86/richgo ********************)
	richgo test -v ./...

install_deps:
	$(info ******************** downloading dependencies ********************)
	go get -v ./...
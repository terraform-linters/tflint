default: build

prepare:
	go mod vendor

test: prepare
	go test -timeout 5m $$(go list ./... | grep -v test-fixtures | grep -v vendor | grep -v aws-sdk-go)

build: test
	mkdir -p dist
	go build -v -o dist/tflint

install: test
	go install

clean:
	rm -rf dist/

code: prepare
	go generate ./...

tools:
	go install github.com/golang/mock/mockgen

.PHONY: prepare test build install clean code tools

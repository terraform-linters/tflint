default: build

prepare:
	go run ./plugin/stub-generator

test: prepare
	go test -timeout 5m $$(go list ./... | grep -v test-fixtures | grep -v stub-generator | grep -v integrationtest | grep -v aws-sdk-go)

build: test
	mkdir -p dist
	go build -v -o dist/tflint

install: test
	go install

e2e: prepare
	go test -timeout 5m ./integrationtest/inspection ./integrationtest/langserver

lint:
	go run golang.org/x/lint/golint --set_exit_status $$(go list ./...)
	go vet ./...

clean:
	rm -rf dist/

code:
	go generate ./...

.PHONY: prepare test build install e2e lint clean code
